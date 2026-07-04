package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"nexus-agent/internal/collector"
	"nexus-agent/internal/config"
	"nexus-agent/internal/httpclient"
	"nexus-agent/internal/devicelimit"
	"nexus-agent/internal/proxy"
)

const (
	heartbeatInterval  = 30 * time.Second
	statsInterval      = 60 * time.Second
	aliveInterval      = 30 * time.Second
	configPullTimeout  = 30 * time.Second
	maxConfigFailures  = 3
	watchCheckInterval      = 5 * time.Second
	deviceLimitCheckInterval = 10 * time.Second
	deviceLimitSyncInterval  = 60 * time.Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Nexus Agent starting...")

	// Load configuration from YAML
	cfgPath := "agent.yaml"
	if len(os.Args) > 1 && os.Args[1] == "-config" && len(os.Args) > 2 {
		cfgPath = os.Args[2]
	} else if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Create shared HTTP client for panel communication
	panelURL := cfg.Panel.URL
	panelToken := cfg.Panel.Token

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start one goroutine per node
	for _, nodeCfg := range cfg.Nodes {
		nodeCfg := nodeCfg // capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			runNode(ctx, panelURL, panelToken, nodeCfg)
		}()
	}

	// Wait for shutdown signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down all nodes...", sig)
	cancel()
	wg.Wait()
	log.Printf("Agent stopped")
}

// runNode manages the full lifecycle of a single proxy node.
func runNode(ctx context.Context, panelURL, panelToken string, nodeCfg config.NodeConfig) {
	prefix := fmt.Sprintf("[node:%d] ", nodeCfg.NodeID)

	// Create panel client with node_id
	client := httpclient.NewClient(panelURL, panelToken, nodeCfg.NodeID)

	// Determine node address
	addr := nodeCfg.Address
	if addr == "" || addr == "auto" {
		addr = "0.0.0.0"
	}

	// Create sing-box manager and stats collector
	sbManager := proxy.New(nodeCfg.Singbox)
	statsCol := collector.New(nodeCfg.Singbox.StatsURL, uint(nodeCfg.NodeID))

	// Pull initial configuration
	log.Printf("%sFetching initial config...", prefix)
	configJSON, usersJSON, routesJSON, err := client.GetConfig()
	if err != nil {
		log.Printf("%sFailed to get initial config: %v", prefix, err)
	} else if strings.TrimSpace(configJSON) == "" || configJSON == "{}" {
		log.Printf("%sPanel returned empty config, waiting for admin to configure...", prefix)
	} else {
		log.Printf("%sConfig received (users=%s, routes=%s)", prefix, usersJSON, routesJSON)
		if err := sbManager.Start(configJSON); err != nil {
			log.Printf("%sFailed to start sing-box: %v", prefix, err)
		} else {
			log.Printf("%ssing-box started successfully", prefix)
		}
	}

	// Start the sing-box watcher goroutine
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go watchSingbox(watchCtx, sbManager, client, prefix)

	// Tickers
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()
	statsTicker := time.NewTicker(statsInterval)
	defer statsTicker.Stop()
	aliveTicker := time.NewTicker(aliveInterval)
	defer aliveTicker.Stop()

	// Device limit enforcement
	deviceLimitEnforcer := devicelimit.New(nodeCfg.Singbox.StatsURL)
	deviceLimitTicker := time.NewTicker(deviceLimitCheckInterval)
	defer deviceLimitTicker.Stop()
	syncLimitsTicker := time.NewTicker(deviceLimitSyncInterval)
	defer syncLimitsTicker.Stop()

	// Initial device limit sync
	if limits, err := client.FetchDeviceLimit(); err == nil {
		deviceLimitEnforcer.UpdateLimits(limits)
		log.Printf("%sDevice limits synced: %d users with limits", prefix, len(limits))
	}

	startTime := time.Now()
	configFailures := 0

	log.Printf("%sRunning: heartbeat=%s, stats=%s, alive=%s", prefix, heartbeatInterval, statsInterval, aliveInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%sShutting down...", prefix)
			if sbManager.IsRunning() {
				if err := sbManager.Stop(); err != nil {
					log.Printf("%sError stopping sing-box: %v", prefix, err)
				}
			}
			return

		case <-heartbeatTicker.C:
			cpu, mem := getSystemStats()
			uptime := uint64(time.Since(startTime).Seconds())

			configChanged, err := client.Heartbeat(cpu, mem, uptime)
			if err != nil {
				log.Printf("%sHeartbeat error: %v", prefix, err)
				continue
			}
			log.Printf("%sHeartbeat OK (cpu=%.1f%%, mem=%.1f%%, uptime=%ds)", prefix, cpu, mem, uptime)

			if configChanged {
				log.Printf("%sConfig change detected, pulling new config...", prefix)
				newConfig, newUsers, newRoutes, err := client.GetConfig()
				if err != nil {
					configFailures++
					log.Printf("%sConfig pull failed (%d/%d): %v", prefix, configFailures, maxConfigFailures, err)
					if configFailures >= maxConfigFailures {
						log.Printf("%sConfig pull failed %d times consecutively, restarting sing-box...", prefix, maxConfigFailures)
						if sbManager.IsRunning() {
							if err := sbManager.Stop(); err != nil {
								log.Printf("%sError stopping sing-box: %v", prefix, err)
							}
						}
						configFailures = 0
					}
					continue
				}
				// Config pull succeeded, reset failure counter
				configFailures = 0

				if strings.TrimSpace(newConfig) == "" || newConfig == "{}" {
					log.Printf("%sNew config is empty, stopping sing-box", prefix)
					if err := sbManager.Stop(); err != nil {
						log.Printf("%sFailed to stop sing-box: %v", prefix, err)
					}
				} else {
					log.Printf("%sRestarting sing-box with new config...", prefix)
					if err := sbManager.Restart(newConfig); err != nil {
						log.Printf("%sFailed to restart sing-box: %v", prefix, err)
					} else {
						log.Printf("%ssing-box restarted successfully (users=%s, routes=%s)", prefix, newUsers, newRoutes)
					}
				}
			}

		case <-statsTicker.C:
			if !sbManager.IsRunning() {
				continue
			}
			entries, err := statsCol.Collect()
			if err != nil {
				log.Printf("%sStats collection error: %v", prefix, err)
				continue
			}
			if len(entries) == 0 {
				continue
			}
			log.Printf("%sCollected traffic stats for %d users", prefix, len(entries))
			if err := client.ReportTraffic(entries); err != nil {
				log.Printf("%sFailed to report traffic: %v", prefix, err)
			}

		case <-aliveTicker.C:
			if !sbManager.IsRunning() {
				continue
			}
			aliveIPs, err := statsCol.CollectAliveIPs()
			if err != nil {
				log.Printf("%sAlive IP collection error: %v", prefix, err)
				continue
			}
			if len(aliveIPs) == 0 {
				continue
			}
			log.Printf("%sReporting alive IPs for %d users", prefix, len(aliveIPs))
			if err := client.ReportAlive(aliveIPs); err != nil {
				log.Printf("%sFailed to report alive IPs: %v", prefix, err)
			}

		case <-deviceLimitTicker.C:
			if !sbManager.IsRunning() || !deviceLimitEnforcer.HasLimits() {
				continue
			}
			closed, err := deviceLimitEnforcer.Enforce()
			if err != nil {
				log.Printf("%sDevice limit enforcement error: %v", prefix, err)
				continue
			}
			if closed > 0 {
				log.Printf("%sDevice limit: closed %d excess connections", prefix, closed)
			}

		case <-syncLimitsTicker.C:
			limits, err := client.FetchDeviceLimit()
			if err != nil {
				log.Printf("%sDevice limit sync error: %v", prefix, err)
				continue
			}
			deviceLimitEnforcer.UpdateLimits(limits)
			if len(limits) > 0 {
				log.Printf("%sDevice limits refreshed: %d users with limits", prefix, len(limits))
			}
		}
	}
}

// watchSingbox monitors the sing-box process and restarts it with exponential backoff
// if it exits unexpectedly. Checks every 5 seconds.
func watchSingbox(ctx context.Context, sbManager *proxy.SingboxManager, client *httpclient.Client, prefix string) {
	var consecutiveFailures int
	maxBackoff := 2 * time.Minute

	ticker := time.NewTicker(watchCheckInterval)
	defer ticker.Stop()

	wasRunning := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			isRunning := sbManager.IsRunning()

			// Detect transition from running to not-running (unexpected exit)
			if wasRunning && !isRunning {
				consecutiveFailures++
				backoff := time.Duration(consecutiveFailures) * 10 * time.Second
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				log.Printf("%ssing-box exited unexpectedly (failure #%d), restarting in %v...", prefix, consecutiveFailures, backoff)

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}

				// Pull latest config before restarting
				configJSON, err := pullConfigWithTimeout(ctx, client, prefix)
				if err != nil {
					log.Printf("%sFailed to pull config for restart: %v", prefix, err)
					wasRunning = false
					continue
				}

				if strings.TrimSpace(configJSON) == "" || configJSON == "{}" {
					log.Printf("%sConfig is empty, not restarting sing-box", prefix)
					wasRunning = false
					continue
				}

				if err := sbManager.Start(configJSON); err != nil {
					log.Printf("%sFailed to restart sing-box: %v", prefix, err)
				} else {
					log.Printf("%ssing-box restarted by watcher (attempt #%d)", prefix, consecutiveFailures)
				}
			}

			// If sing-box is running again, reset failure counter
			if isRunning && consecutiveFailures > 0 {
				consecutiveFailures = 0
			}

			wasRunning = isRunning
		}
	}
}

// pullConfigWithTimeout pulls the config from the panel with a 30-second timeout.
func pullConfigWithTimeout(ctx context.Context, client *httpclient.Client, prefix string) (string, error) {
	type result struct {
		configJSON string
		err        error
	}

	ch := make(chan result, 1)
	go func() {
		configJSON, _, _, err := client.GetConfig()
		if err != nil {
			ch <- result{"", err}
			return
		}
		ch <- result{configJSON, nil}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(configPullTimeout):
		return "", fmt.Errorf("config pull timed out after %s", configPullTimeout)
	case r := <-ch:
		return r.configJSON, r.err
	}
}

// getSystemStats returns CPU usage percentage and memory usage percentage.
func getSystemStats() (cpu float64, mem float64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if m.Sys > 0 {
		mem = float64(m.Alloc) / float64(m.Sys) * 100
	}

	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	if numCPU > 0 {
		cpu = float64(numGoroutines) / float64(numCPU) * 10
		if cpu > 100 {
			cpu = 100
		}
	}

	return cpu, mem
}
