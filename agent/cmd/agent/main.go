package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"nexus-agent/internal/cert"
	"nexus-agent/internal/collector"
	"nexus-agent/internal/config"
	"nexus-agent/internal/devicelimit"
	"nexus-agent/internal/httpclient"
	"nexus-agent/internal/kernel"
	"nexus-agent/internal/logger"
	"nexus-agent/internal/loglevel"
	"nexus-agent/internal/machine"
	"nexus-agent/internal/proxy"
	"nexus-agent/internal/wsclient"
)

const (
	heartbeatInterval        = 30 * time.Second
	statsInterval            = 60 * time.Second
	aliveInterval            = 30 * time.Second
	configPullTimeout        = 30 * time.Second
	maxConfigFailures        = 3
	watchCheckInterval       = 5 * time.Second
	deviceLimitCheckInterval = 10 * time.Second
	deviceLimitSyncInterval  = 60 * time.Second
)

// sanitizeSingboxConfig cleans deprecated / invalid DNS fields for sing-box 1.12+.
func sanitizeSingboxConfig(configJSON string) string {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Printf("Warning: failed to parse config JSON for cleanup: %v", err)
		return configJSON
	}

	changed := false
	dns, ok := config["dns"].(map[string]interface{})
	if !ok {
		return configJSON
	}

	// Remove legacy top-level fakeip block
	if _, hasFakeIP := dns["fakeip"]; hasFakeIP {
		delete(dns, "fakeip")
		changed = true
		log.Printf("Removed legacy dns.fakeip for sing-box 1.12+ compatibility")
	}

	// Sanitize servers list
	if servers, ok := dns["servers"].([]interface{}); ok {
		newServers := make([]interface{}, 0, len(servers))
		for _, raw := range servers {
			server, ok := raw.(map[string]interface{})
			if !ok {
				newServers = append(newServers, raw)
				continue
			}

			// Drop incomplete fakeip servers without ranges
			addr, _ := server["address"].(string)
			stype, _ := server["type"].(string)
			if addr == "fakeip" || stype == "fakeip" {
				inet4, _ := server["inet4_range"].(string)
				inet6, _ := server["inet6_range"].(string)
				if inet4 == "" && inet6 == "" {
					changed = true
					log.Printf("Removed incomplete fakeip DNS server from config")
					continue
				}
			}

			// Convert legacy {"address":"https://1.1.1.1/dns-query"} to new format
			if addr != "" && stype == "" {
				if strings.HasPrefix(addr, "https://") {
					server["type"] = "https"
					host := strings.TrimPrefix(addr, "https://")
					host = strings.Split(host, "/")[0]
					server["server"] = host
					delete(server, "address")
					changed = true
				} else if strings.HasPrefix(addr, "tls://") {
					server["type"] = "tls"
					server["server"] = strings.TrimPrefix(addr, "tls://")
					delete(server, "address")
					changed = true
				} else if addr == "local" {
					server["type"] = "local"
					delete(server, "address")
					changed = true
				}
			}

			newServers = append(newServers, server)
		}
		dns["servers"] = newServers
	}

	if !changed {
		return configJSON
	}

	cleanedBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to re-marshal cleaned config: %v", err)
		return configJSON
	}
	return string(cleanedBytes)
}

// removeLegacyFakeIP is kept for compatibility and now uses full sanitizer.
func removeLegacyFakeIP(configJSON string) string {
	return sanitizeSingboxConfig(configJSON)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Nexus Agent starting...")

	// Load configuration from YAML
	// Accept: -c path | -config path | --config path | bare path
	cfgPath := "config.yml"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch {
		case (arg == "-c" || arg == "-config" || arg == "--config") && len(os.Args) > 2:
			cfgPath = os.Args[2]
		case arg == "-v" || arg == "--version":
			fmt.Printf("nexus-agent dev\n")
			os.Exit(0)
		case !strings.HasPrefix(arg, "-"):
			cfgPath = arg
		default:
			// Unknown flag style: try next arg as path if present
			if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
				cfgPath = os.Args[2]
			}
		}
	}

	// Resolve config path to absolute so relative working_dir is correct under any CWD.
	if absCfg, err := filepath.Abs(cfgPath); err == nil {
		cfgPath = absCfg
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config %s: %v", cfgPath, err)
	}

	// If working_dir is relative/empty, default it to the config file directory.
	cfgDir := filepath.Dir(cfgPath)
	if err := os.Chdir(cfgDir); err != nil {
		log.Printf("Warning: failed to chdir to %s: %v", cfgDir, err)
	} else {
		log.Printf("Working directory: %s", cfgDir)
	}

	// Configure log level from config
	loglevel.SetLevel(cfg.Log.Level)
	log.Printf("Log level: %s", loglevel.GetLevel())

// Configure file logging if log_file is set
		if cfg.Log.LogFile != "" {
			maxSize := cfg.Log.LogMaxSize
			rotator, err := logger.NewRotatingFileWriter(cfg.Log.LogFile, maxSize, 0)
			if err != nil {
				log.Fatalf("Failed to open log file: %v", err)
			}
			log.SetOutput(rotator)
			log.Printf("Logging to file: %s (max_size=%d)", cfg.Log.LogFile, maxSize)
		}

		// Start health check HTTP server if port is configured
		if cfg.HealthPort > 0 {
			startHealthServer(cfg.HealthPort)
		}

		// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Machine mode: one orchestrator manages all nodes dynamically
	if cfg.IsMachineMode() {
		log.Printf("Machine mode enabled (machine_id=%d)", cfg.Machine.ID)
		orch := machine.New(&cfg)
		if err := orch.Run(ctx); err != nil {
			log.Fatalf("Machine orchestrator failed: %v", err)
		}
		// Wait for shutdown signal
		<-sigCh
		log.Printf("Received signal, shutting down...")
		cancel()
		return
	}

	// Normalize nodes from config for traditional mode
	nodes, err := cfg.NormalizeNodes()
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Printf("Loaded %d node(s) from config", len(nodes))

	var wg sync.WaitGroup

	// Start one goroutine per node
	for _, node := range nodes {
		node := node // capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			runNode(ctx, node)
		}()
	}

	// Wait for shutdown signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down all nodes...", sig)
	cancel()
	wg.Wait()
	log.Printf("Agent stopped")
}

func nodeKernelConfig(panelCfg *kernel.NodeConfigFromPanel, localCfg config.NormalizedNode) kernel.NodeConfig {
	cfg := panelCfg.ToNodeConfig()
	cfg.StatsPort = localCfg.Singbox.StatsPort
	return cfg
}

// startSingbox starts or restarts sing-box with the given node config.
func startSingbox(sbManager *proxy.SingboxManager, nodeConfig *kernel.NodeConfigFromPanel, nodeCfg config.NormalizedNode, users []kernel.User, prefix string) error {
	if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
		log.Printf("%sStarting sing-box with raw config_json (json mode)", prefix)
		cleanedConfig := sanitizeSingboxConfig(nodeConfig.ConfigJSON)
		return sbManager.Restart(cleanedConfig)
	}
	configJSON, err := kernel.GenerateSingboxConfig(nodeKernelConfig(nodeConfig, nodeCfg), users)
	if err != nil {
		return fmt.Errorf("generate config: %w", err)
	}
	configJSON = sanitizeSingboxConfig(configJSON)
	log.Printf("%sStarting sing-box (protocol=%s, port=%d, %d users)", prefix, nodeConfig.Protocol, nodeConfig.ServerPort, len(users))
	return sbManager.Restart(configJSON)
}

// hotReloadOrRestart tries a hot config reload via PUT /configs first.
// Falls back to a full restart if hot reload fails or sing-box is not running.
func hotReloadOrRestart(sbManager *proxy.SingboxManager, nodeConfig *kernel.NodeConfigFromPanel, nodeCfg config.NormalizedNode, users []kernel.User, prefix string) error {
	// Generate config
	var configJSON string
	if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
		configJSON = sanitizeSingboxConfig(nodeConfig.ConfigJSON)
	} else {
		var err error
		configJSON, err = kernel.GenerateSingboxConfig(nodeKernelConfig(nodeConfig, nodeCfg), users)
		if err != nil {
			return fmt.Errorf("generate config: %w", err)
		}
		configJSON = sanitizeSingboxConfig(configJSON)
	}

	// Try hot reload if sing-box is running
	if sbManager.IsRunning() {
		if err := sbManager.ReloadConfig(configJSON); err != nil {
			log.Printf("%sHot reload failed (%v), falling back to full restart", prefix, err)
			return sbManager.Restart(configJSON)
		}
		log.Printf("%sConfig hot-reloaded (protocol=%s, %d users)", prefix, nodeConfig.Protocol, len(users))
		return nil
	}

	// Not running, do a full start
	log.Printf("%sStarting sing-box (protocol=%s, port=%d, %d users)", prefix, nodeConfig.Protocol, nodeConfig.ServerPort, len(users))
	return sbManager.Start(configJSON)
}


func userUUIDs(users []kernel.User) []string {
	out := make([]string, 0, len(users))
	for _, u := range users {
		if u.UUID != "" {
			out = append(out, u.UUID)
		}
	}
	return out
}

func applyKnownUsers(statsCol *collector.StatsCollector, enforcer *devicelimit.Enforcer, users []kernel.User) {
	uuids := userUUIDs(users)
	if statsCol != nil {
		statsCol.SetKnownUsers(uuids)
	}
	if enforcer != nil {
		enforcer.SetKnownUsers(uuids)
	}
}

// runNode manages the full lifecycle of a single proxy node.
func runNode(ctx context.Context, nodeCfg config.NormalizedNode) {
	prefix := fmt.Sprintf("[node:%d] ", nodeCfg.NodeID)

	// Create panel client with node_id
	client := httpclient.NewClient(nodeCfg.PanelURL, nodeCfg.PanelToken, fmt.Sprintf("%d", nodeCfg.NodeID))

	// Create sing-box manager and stats collector
	sbManager := proxy.New(nodeCfg.Singbox)
	statsCol := collector.New(nodeCfg.Singbox.StatsURL, fmt.Sprintf("%d", nodeCfg.NodeID))

	// Perform handshake
	log.Printf("%sPerforming handshake...", prefix)
	handshake, err := client.Handshake()
	if err != nil {
		log.Printf("%sHandshake failed: %v, using defaults", prefix, err)
	} else {
		log.Printf("%sHandshake OK (push_interval=%ds, pull_interval=%ds)",
			prefix, handshake.Settings.PushInterval, handshake.Settings.PullInterval)
	}

	// Pull initial configuration
	log.Printf("%sFetching initial config...", prefix)
	nodeConfig, users, err := pullConfigWithUsers(client, prefix)
if err != nil {
			log.Printf("%sFailed to get initial config: %v", prefix, err)
		} else if nodeConfig == nil {
			log.Printf("%sPanel returned empty config, waiting for admin to configure...", prefix)
} else if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
			// JSON mode: use the raw config_json directly
			log.Printf("%sUsing raw config_json (json mode)", prefix)
			cleanedConfig := sanitizeSingboxConfig(nodeConfig.ConfigJSON)
			if err := sbManager.Start(cleanedConfig); err != nil {
				log.Printf("%sFailed to start sing-box: %v", prefix, err)
			} else {
				log.Printf("%ssing-box started successfully (json mode)", prefix)
			}
		} else {
			// Auto mode: generate sing-box config from node parameters
			configJSON, err := kernel.GenerateSingboxConfig(nodeKernelConfig(nodeConfig, nodeCfg), users)
			if err != nil {
				log.Printf("%sFailed to generate sing-box config: %v", prefix, err)
			} else {
				configJSON = sanitizeSingboxConfig(configJSON)
				log.Printf("%sConfig generated (protocol=%s, port=%d, %d users)",
					prefix, nodeConfig.Protocol, nodeConfig.ServerPort, len(users))
				if err := sbManager.Start(configJSON); err != nil {
					log.Printf("%sFailed to start sing-box: %v", prefix, err)
				} else {
					log.Printf("%ssing-box started successfully", prefix)
				}
			}
		}

		// WS update channel: WS handlers send data here, main loop applies it
		type wsUpdate struct {
			config *kernel.NodeConfigFromPanel
			users  []kernel.User
		}
		wsUpdateCh := make(chan wsUpdate, 1)
		var wsClient *wsclient.Client

		// Connect to panel WebSocket for real-time commands
		if handshake != nil && handshake.WebSocket.Enabled && handshake.WebSocket.WSURL != "" {
			wsClient = wsclient.NewClient(handshake.WebSocket.WSURL, nodeCfg.PanelToken, fmt.Sprintf("%d", nodeCfg.NodeID))
			wsClient.RegisterHandler("restart", func(cmd wsclient.Command) error {
				log.Printf("%sReceived restart command from panel", prefix)
				if err := sbManager.Stop(); err != nil {
					log.Printf("%sError stopping sing-box: %v", prefix, err)
				}
				// Pull fresh config and restart
				newNodeConfig, newUsers, err := pullConfigWithUsers(client, prefix)
				if err != nil {
					return fmt.Errorf("config pull failed: %w", err)
				}
				if newNodeConfig == nil {
					return nil
				}
				return startSingbox(sbManager, newNodeConfig, nodeCfg, newUsers, prefix)
			})
wsClient.RegisterHandler("reload", func(cmd wsclient.Command) error {
					log.Printf("%sReceived reload command from panel", prefix)
					newNodeConfig, newUsers, err := pullConfigWithUsers(client, prefix)
					if err != nil {
						return fmt.Errorf("config pull failed: %w", err)
					}
					if newNodeConfig == nil {
						return nil
					}
					return hotReloadOrRestart(sbManager, newNodeConfig, nodeCfg, newUsers, prefix)
				})

				// update: panel pushes a new binary URL — download, replace, restart
				wsClient.RegisterHandler("update", func(cmd wsclient.Command) error {
					log.Printf("%sReceived update command from panel", prefix)
					var data struct {
						DownloadURL string `json:"download_url"`
					}
					if err := json.Unmarshal(cmd.Data, &data); err != nil {
						return fmt.Errorf("parse update command: %w", err)
					}
					if data.DownloadURL == "" {
						return fmt.Errorf("update command missing download_url")
					}

					// Download new binary
					tmpPath := nodeCfg.Singbox.BinaryPath + ".tmp"
					if err := downloadFile(data.DownloadURL, tmpPath); err != nil {
						return fmt.Errorf("download update: %w", err)
					}

					// Make it executable
					if err := os.Chmod(tmpPath, 0755); err != nil {
						os.Remove(tmpPath)
						return fmt.Errorf("chmod update: %w", err)
					}

					// Stop sing-box gracefully
					if err := sbManager.Stop(); err != nil {
						log.Printf("%sError stopping sing-box for update: %v", prefix, err)
					}

					// Replace binary
					if err := os.Rename(tmpPath, nodeCfg.Singbox.BinaryPath); err != nil {
						os.Remove(tmpPath)
						return fmt.Errorf("replace binary: %w", err)
					}

					log.Printf("%sBinary updated, restarting sing-box...", prefix)

					// Pull fresh config and restart
					newNodeConfig, newUsers, err := pullConfigWithUsers(client, prefix)
					if err != nil {
						return fmt.Errorf("config pull after update failed: %w", err)
					}
					if newNodeConfig == nil {
						return nil
					}
					return startSingbox(sbManager, newNodeConfig, nodeCfg, newUsers, prefix)
				})

			// sync.config: panel pushes full node config via WS
			wsClient.RegisterHandler("sync.config", func(cmd wsclient.Command) error {
				log.Printf("%sReceived sync.config from panel", prefix)
				var nodeConfig httpclient.NodeConfigResponse
				if err := json.Unmarshal(cmd.Data, &nodeConfig); err != nil {
					return fmt.Errorf("parse sync.config: %w", err)
				}
				// Convert to kernel type
				certConfig := convertCertConfig(nodeConfig.CertConfig)
				certPEM, keyPEM, _ := loadCertMaterial(certConfig)
				kCfg := &kernel.NodeConfigFromPanel{
					ConfigMode:        nodeConfig.ConfigMode,
					ConfigJSON:        nodeConfig.ConfigJSON,
					NodeID:            nodeConfig.NodeID,
					Protocol:          nodeConfig.Protocol,
					ListenIP:          nodeConfig.ListenIP,
					ServerPort:        nodeConfig.ServerPort,
					Network:           nodeConfig.Network,
					NetworkSettings:   nodeConfig.NetworkSettings,
					KernelType:        nodeConfig.KernelType,
					CertConfig:        certConfig,
					CustomOutbounds:   convertCustomOutbounds(nodeConfig.CustomOutbounds),
					CertPEM:           certPEM,
					KeyPEM:            keyPEM,
					TLS:               nodeConfig.TLS,
					Flow:              nodeConfig.Flow,
					TLSSettings:       nodeConfig.TLSSettings,
					ServerName:        nodeConfig.ServerName,
					UpMbps:            nodeConfig.UpMbps,
					DownMbps:          nodeConfig.DownMbps,
					ObfsPassword:      nodeConfig.ObfsPassword,
					CongestionControl: nodeConfig.CongestionControl,
					BaseConfig: kernel.BaseConfig{
						PushInterval: nodeConfig.BaseConfig.PushInterval,
						PullInterval: nodeConfig.BaseConfig.PullInterval,
					},
				}
				if len(nodeConfig.Routes) > 0 {
					kCfg.Routes = make([]kernel.RouteRule, len(nodeConfig.Routes))
					for i, r := range nodeConfig.Routes {
						kCfg.Routes[i] = kernel.RouteRule{
							ID: r.ID, Match: r.Match, MatchRule: r.MatchRule,
							Action: r.Action, ActionValue: r.ActionValue, ActionRule: r.ActionRule,
						}
					}
				}
				select {
				case wsUpdateCh <- wsUpdate{config: kCfg}:
				default:
				}
				return nil
			})

			// sync.users: panel pushes full user list via WS
			wsClient.RegisterHandler("sync.users", func(cmd wsclient.Command) error {
				log.Printf("%sReceived sync.users from panel", prefix)
				var resp httpclient.UsersResponse
				if err := json.Unmarshal(cmd.Data, &resp); err != nil {
					return fmt.Errorf("parse sync.users: %w", err)
				}
				users := make([]kernel.User, len(resp.Users))
				for i, u := range resp.Users {
					users[i] = kernel.User{ID: u.ID, UUID: u.UUID, SpeedLimit: u.SpeedLimit, DeviceLimit: u.DeviceLimit}
				}
				select {
				case wsUpdateCh <- wsUpdate{users: users}:
				default:
				}
				return nil
			})

			// On reconnect, pull fresh config and hot-reload to catch changes made while disconnected
			wsClient.OnReconnect = func() {
				log.Printf("%sWS reconnected, syncing config...", prefix)
				newNodeConfig, newUsers, err := pullConfigWithUsers(client, prefix)
				if err != nil {
					log.Printf("%sPost-reconnect config sync failed: %v", prefix, err)
					return
				}
				if newNodeConfig != nil {
					if err := hotReloadOrRestart(sbManager, newNodeConfig, nodeCfg, newUsers, prefix); err != nil {
						log.Printf("%sPost-reconnect reload failed: %v", prefix, err)
					}
				}
			}

			if err := wsClient.Connect(); err != nil {
				log.Printf("%sWebSocket connection failed: %v (continuing with HTTP polling)", prefix, err)
			} else {
				go wsClient.ReconnectLoop()
				log.Printf("%sWebSocket connected for real-time commands", prefix)
			}
		} else {
			log.Printf("%sWebSocket not available from handshake, using HTTP polling only", prefix)
		}

		// Start the sing-box watcher goroutine
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go watchSingbox(watchCtx, sbManager, client, prefix, nodeCfg)

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
	pendingTraffic := make(map[string][2]int64) // buffered traffic from failed reports
	var lastNodeConfig *kernel.NodeConfigFromPanel // cached config for WS push
	var lastUsers []kernel.User                    // cached users for WS push
	if users != nil {
		lastUsers = users
		applyKnownUsers(statsCol, deviceLimitEnforcer, users)
	}

	// Use handshake intervals if available
	pushInterval := heartbeatInterval
	if handshake != nil {
		if handshake.Settings.PushInterval > 0 {
			pushInterval = time.Duration(handshake.Settings.PushInterval) * time.Second
		}
		heartbeatTicker.Reset(pushInterval)
	}

	log.Printf("%sRunning: heartbeat=%s, stats=%s, alive=%s", prefix, pushInterval, statsInterval, aliveInterval)

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

			configChanged, newPullInterval, err := client.Heartbeat(cpu, mem, uptime)
			if err != nil {
				log.Printf("%sHeartbeat error: %v", prefix, err)
				continue
			}
			log.Printf("%sHeartbeat OK (cpu=%.1f%%, mem=%.1f%%, uptime=%ds)", prefix, cpu, mem, uptime)

			// Adjust heartbeat interval based on panel's pull_interval
			if newPullInterval > 0 {
				heartbeatTicker.Reset(time.Duration(newPullInterval) * time.Second)
			}

			// Skip config change detection when WS is connected (config is pushed via WS)
			if wsClient != nil && wsClient.IsConnected() {
				continue
			}

			if configChanged {
				log.Printf("%sConfig change detected, pulling new config...", prefix)
				newNodeConfig, newUsers, err := pullConfigWithUsers(client, prefix)
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
				configFailures = 0
				if newNodeConfig == nil {
					log.Printf("%sNew config is empty, stopping sing-box", prefix)
					if err := sbManager.Stop(); err != nil {
						log.Printf("%sFailed to stop sing-box: %v", prefix, err)
					}
} else {
						if err := hotReloadOrRestart(sbManager, newNodeConfig, nodeCfg, newUsers, prefix); err != nil {
							log.Printf("%sFailed to apply config: %v", prefix, err)
						} else {
							lastUsers = newUsers
							applyKnownUsers(statsCol, deviceLimitEnforcer, newUsers)
						}
					}
				}

			case <-statsTicker.C:
			if !sbManager.IsRunning() {
				continue
			}
			trafficData, err := statsCol.CollectXboard()
			if err != nil {
				log.Printf("%sStats collection error: %v", prefix, err)
				continue
			}

			// Merge new traffic with pending buffer
			for uuid, delta := range trafficData {
				pendingTraffic[uuid] = [2]int64{
					pendingTraffic[uuid][0] + delta[0],
					pendingTraffic[uuid][1] + delta[1],
				}
			}

// Cap pending buffer size to prevent memory leak
				const maxPendingUsers = 5000
				if len(pendingTraffic) > maxPendingUsers {
					// Drop oldest half to make room — more graceful than full reset
					dropCount := len(pendingTraffic) - maxPendingUsers/2
					newMap := make(map[string][2]int64, maxPendingUsers/2)
					count := 0
					for k, v := range pendingTraffic {
						if count >= dropCount {
							newMap[k] = v
						}
						count++
					}
					log.Printf("%sPending traffic buffer too large (%d), dropping %d oldest entries", prefix, len(pendingTraffic), dropCount)
					pendingTraffic = newMap
				}

			aliveIPs, err := statsCol.CollectAliveIPs()
			if err != nil {
				log.Printf("%sAlive IP collection error: %v", prefix, err)
				aliveIPs = map[string][]string{}
			}
			cpu, mem := getSystemStats()
			log.Printf("%sReporting consolidated stats: traffic=%d (+%d pending) alive=%d", prefix, len(trafficData), len(pendingTraffic), len(aliveIPs))
			if err := client.Report(pendingTraffic, aliveIPs, cpu, mem, 0); err != nil {
				log.Printf("%sFailed to report consolidated stats (will retry next cycle): %v", prefix, err)
			} else {
				// Report succeeded, clear pending buffer
				pendingTraffic = make(map[string][2]int64)
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

case update := <-wsUpdateCh:
				// WS pushed config or users — apply via hot reload
				if update.config != nil {
					lastNodeConfig = update.config
				}
				if update.users != nil {
					lastUsers = update.users
					applyKnownUsers(statsCol, deviceLimitEnforcer, update.users)
				}
				if lastNodeConfig != nil {
					if err := hotReloadOrRestart(sbManager, lastNodeConfig, nodeCfg, lastUsers, prefix); err != nil {
						log.Printf("%sWS sync apply failed: %v", prefix, err)
					} else {
						log.Printf("%sConfig applied from WS push", prefix)
						applyKnownUsers(statsCol, deviceLimitEnforcer, lastUsers)
					}
				}
			}
	}
}

// pullConfigWithUsers fetches node config and users from the panel.
func pullConfigWithUsers(client *httpclient.Client, prefix string) (*kernel.NodeConfigFromPanel, []kernel.User, error) {
	type result struct {
		config *kernel.NodeConfigFromPanel
		users  []kernel.User
		err    error
	}

	ch := make(chan result, 1)
	go func() {
		nodeConfig, err := client.GetConfig()
		if err != nil {
			ch <- result{nil, nil, err}
			return
		}

		usersInfo, err := client.GetUsers()
		if err != nil {
			ch <- result{nil, nil, err}
			return
		}

		certConfig := convertCertConfig(nodeConfig.CertConfig)
		certPEM, keyPEM, err := loadCertMaterial(certConfig)
		if err != nil {
			ch <- result{nil, nil, err}
			return
		}

// Convert to kernel format
			config := &kernel.NodeConfigFromPanel{
				ConfigMode:        nodeConfig.ConfigMode,
				ConfigJSON:        nodeConfig.ConfigJSON,
				NodeID:            nodeConfig.NodeID,
				Protocol:          nodeConfig.Protocol,
				ListenIP:          nodeConfig.ListenIP,
				ServerPort:        nodeConfig.ServerPort,
				Network:           nodeConfig.Network,
				NetworkSettings:   nodeConfig.NetworkSettings,
				KernelType:        nodeConfig.KernelType,
				CertConfig:        certConfig,
				CustomOutbounds:   convertCustomOutbounds(nodeConfig.CustomOutbounds),
				CertPEM:           certPEM,
				KeyPEM:            keyPEM,
				TLS:               nodeConfig.TLS,
				Flow:              nodeConfig.Flow,
				TLSSettings:       nodeConfig.TLSSettings,
				ServerName:        nodeConfig.ServerName,
				UpMbps:            nodeConfig.UpMbps,
				DownMbps:          nodeConfig.DownMbps,
				ObfsPassword:      nodeConfig.ObfsPassword,
				CongestionControl: nodeConfig.CongestionControl,
				BaseConfig: kernel.BaseConfig{
				PushInterval: nodeConfig.BaseConfig.PushInterval,
				PullInterval: nodeConfig.BaseConfig.PullInterval,
			},
		}

		// Convert routes
		if len(nodeConfig.Routes) > 0 {
			config.Routes = make([]kernel.RouteRule, len(nodeConfig.Routes))
			for i, r := range nodeConfig.Routes {
				config.Routes[i] = kernel.RouteRule{
					ID:          r.ID,
					Match:       r.Match,
					MatchRule:   r.MatchRule,
					Action:      r.Action,
					ActionValue: r.ActionValue,
					ActionRule:  r.ActionRule,
				}
			}
		}

		// Convert to kernel.User format
		users := make([]kernel.User, len(usersInfo))
		for i, u := range usersInfo {
			users[i] = kernel.User{
				ID:          u.ID,
				UUID:        u.UUID,
				SpeedLimit:  u.SpeedLimit,
				DeviceLimit: u.DeviceLimit,
			}
		}

		ch <- result{config, users, nil}
	}()

	select {
	case <-time.After(configPullTimeout):
		return nil, nil, fmt.Errorf("config pull timed out after %s", configPullTimeout)
	case r := <-ch:
		return r.config, r.users, r.err
	}
}

func convertCertConfig(in httpclient.CertConfig) kernel.CertConfig {
	return kernel.CertConfig{
		CertMode:    in.CertMode,
		Domain:      in.Domain,
		Email:       in.Email,
		DNSProvider: in.DNSProvider,
		DNSEnv:      in.DNSEnv,
		HTTPPort:    in.HTTPPort,
		CertFile:    in.CertFile,
		KeyFile:     in.KeyFile,
		CertContent: in.CertContent,
		KeyContent:  in.KeyContent,
		CertDir:     in.CertDir,
	}
}

func loadCertMaterial(cfg kernel.CertConfig) (string, string, error) {
	if cfg.CertMode == "" && cfg.CertFile == "" && cfg.CertContent == "" {
		return "", "", nil
	}
	manager := cert.NewManager(cert.Config{
		CertMode:    cfg.CertMode,
		Domain:      cfg.Domain,
		Email:       cfg.Email,
		DNSProvider: cfg.DNSProvider,
		DNSEnv:      cfg.DNSEnv,
		HTTPPort:    cfg.HTTPPort,
		CertFile:    cfg.CertFile,
		KeyFile:     cfg.KeyFile,
		CertContent: cfg.CertContent,
		KeyContent:  cfg.KeyContent,
		CertDir:     cfg.CertDir,
	})
	ctx, cancel := context.WithTimeout(context.Background(), configPullTimeout)
	defer cancel()
	if err := manager.Start(ctx); err != nil {
		return "", "", fmt.Errorf("load cert material: %w", err)
	}
	material := manager.Material()
	return string(material.CertPEM), string(material.KeyPEM), nil
}

func convertCustomOutbounds(in []httpclient.CustomOutbound) []kernel.CustomOutbound {
	out := make([]kernel.CustomOutbound, len(in))
	for i, co := range in {
		out[i] = kernel.CustomOutbound{
			Tag:      co.Tag,
			Protocol: co.Protocol,
			Settings: co.Settings,
			ProxyTag: co.ProxyTag,
		}
	}
	return out
}

// watchSingbox monitors the sing-box process and restarts it with exponential backoff
// if it exits unexpectedly or never started successfully.
func watchSingbox(ctx context.Context, sbManager *proxy.SingboxManager, client *httpclient.Client, prefix string, nodeCfg config.NormalizedNode) {
	var consecutiveFailures int
	maxBackoff := 2 * time.Minute

	ticker := time.NewTicker(watchCheckInterval)
	defer ticker.Stop()

	wasRunning := false
	lastStartAttempt := time.Time{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			isRunning := sbManager.IsRunning()

			// Unexpected exit, or never successfully running after initial attempt.
			needRestart := (wasRunning && !isRunning) || (!isRunning && time.Since(lastStartAttempt) > 15*time.Second)
			if !needRestart {
				if isRunning && consecutiveFailures > 0 {
					consecutiveFailures = 0
				}
				wasRunning = isRunning
				continue
			}

			consecutiveFailures++
			backoff := time.Duration(consecutiveFailures) * 5 * time.Second
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			if wasRunning {
				log.Printf("%ssing-box exited unexpectedly (failure #%d), restarting in %v...", prefix, consecutiveFailures, backoff)
			} else {
				log.Printf("%ssing-box not running (attempt #%d), starting in %v...", prefix, consecutiveFailures, backoff)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}

			// Pull latest config before restarting
			nodeConfig, users, err := pullConfigWithUsers(client, prefix)
			if err != nil {
				log.Printf("%sFailed to pull config for restart: %v", prefix, err)
				wasRunning = false
				lastStartAttempt = time.Now()
				continue
			}

			if nodeConfig == nil {
				log.Printf("%sConfig is empty, not restarting sing-box", prefix)
				wasRunning = false
				lastStartAttempt = time.Now()
				continue
			}

			// Generate sing-box config from node parameters
			lastStartAttempt = time.Now()
			if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
				log.Printf("%sRestarting sing-box with raw config_json (json mode)...", prefix)
				cleanedConfig := sanitizeSingboxConfig(nodeConfig.ConfigJSON)
				if err := sbManager.Start(cleanedConfig); err != nil {
					log.Printf("%sFailed to restart sing-box: %v", prefix, err)
					wasRunning = false
				} else {
					log.Printf("%ssing-box restarted by watcher (json mode)", prefix)
					wasRunning = true
					consecutiveFailures = 0
				}
			} else {
				configJSON, err := kernel.GenerateSingboxConfig(nodeKernelConfig(nodeConfig, nodeCfg), users)
				if err != nil {
					log.Printf("%sFailed to generate config: %v", prefix, err)
					wasRunning = false
					continue
				}
				configJSON = sanitizeSingboxConfig(configJSON)

				if err := sbManager.Start(configJSON); err != nil {
					log.Printf("%sFailed to restart sing-box: %v", prefix, err)
					wasRunning = false
				} else {
					log.Printf("%ssing-box restarted by watcher (attempt #%d)", prefix, consecutiveFailures)
					wasRunning = true
					consecutiveFailures = 0
				}
			}
		}
	}
}

// startHealthServer starts a minimal HTTP server that responds to /health.
func startHealthServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})
	go func() {
		addr := fmt.Sprintf(":%d", port)
		log.Printf("Health check server listening on %s/health", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()
}

// downloadFile downloads a URL to a local path.
func downloadFile(url, path string) error {
	log.Printf("Downloading %s -> %s", url, path)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	log.Printf("Downloaded %d bytes", written)
	return nil
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
