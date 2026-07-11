// Package machine implements the machine-mode orchestrator.
//
// In machine mode, a single agent process manages all nodes bound to a
// machine on the panel. Nodes are discovered dynamically via the machine API.
package machine

import (
		"context"
		"encoding/json"
		"fmt"
		"io"
		"log"
		"net/http"
		"os"
		"runtime"
		"sync"
		"time"

	"nexus-agent/internal/cert"
	"nexus-agent/internal/collector"
	"nexus-agent/internal/config"
	"nexus-agent/internal/devicelimit"
	"nexus-agent/internal/httpclient"
	"nexus-agent/internal/kernel"
	"nexus-agent/internal/proxy"
	"nexus-agent/internal/system"
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
	loadReportInterval       = 60 * time.Second
)

// nodeHandle tracks a running node service.
type nodeHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Orchestrator manages all nodes bound to a panel machine.
type Orchestrator struct {
	cfg    *config.Config
	client *httpclient.Client

	mu      sync.Mutex
	nodes   map[int]*nodeHandle
	nodeCfg map[int]config.NormalizedNode

	wsClient     *wsclient.Client
	pullInterval  time.Duration
	pushInterval  time.Duration
	loadMonitor   *system.Monitor
}

// New creates a machine orchestrator from the given config.
func New(cfg *config.Config) *Orchestrator {
	panelURL := cfg.Machine.PanelURL
	if panelURL == "" {
		panelURL = cfg.Panel.URL
	}
	client := httpclient.NewClient(panelURL, cfg.Machine.Token, "")
	client.SetMachineID(cfg.Machine.ID)

	return &Orchestrator{
		cfg:         cfg,
		client:      client,
		nodes:       make(map[int]*nodeHandle),
		nodeCfg:     make(map[int]config.NormalizedNode),
		loadMonitor: system.NewMonitor(),
	}
}

// Run starts the machine orchestrator. Blocks until ctx is cancelled.
func (o *Orchestrator) Run(ctx context.Context) error {
	log.Printf("[machine] Starting machine mode (machine_id=%d)", o.cfg.Machine.ID)

	// Discover initial nodes
	nodesResp, err := o.client.GetMachineNodes()
	if err != nil {
		return fmt.Errorf("initial node discovery: %w", err)
	}

	o.pullInterval = time.Duration(nodesResp.PullInterval) * time.Second
	if o.pullInterval < 30*time.Second {
		o.pullInterval = 60 * time.Second
	}
	o.pushInterval = time.Duration(nodesResp.PushInterval) * time.Second
	if o.pushInterval < 10*time.Second {
		o.pushInterval = 60 * time.Second
	}

	log.Printf("[machine] Discovered %d nodes", len(nodesResp.Nodes))

	// Store node configs
	for _, n := range nodesResp.Nodes {
		o.nodeCfg[int(n.ID)] = config.NormalizedNode{
			NodeID:     int(n.ID),
			PanelURL:   o.cfg.Panel.URL,
			PanelToken: o.cfg.Machine.Token,
			MachineID:  o.cfg.Machine.ID,
			Singbox:    o.cfg.Singbox,
		}
	}

	// Start initial nodes
	for _, n := range nodesResp.Nodes {
		o.startNode(ctx, int(n.ID))
	}

	// Connect WebSocket for real-time commands
	o.connectWS(ctx)

	// Discovery ticker (poll for node list changes)
	discoveryTicker := time.NewTicker(o.pullInterval)
	defer discoveryTicker.Stop()
heartbeatTicker := time.NewTicker(o.pushInterval)
		defer heartbeatTicker.Stop()
		loadTicker := time.NewTicker(loadReportInterval)
		defer loadTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			o.stopAll()
			return nil
		case <-discoveryTicker.C:
			o.rediscover(ctx)
		case <-heartbeatTicker.C:
			o.client.MachineHeartbeat(o.cfg.Machine.ID)
			case <-loadTicker.C:
				load := o.loadMonitor.Collect()
				o.client.MachineReportLoad(o.cfg.Machine.ID, httpclient.LoadData{
					CPU:         load.CPU,
					MemTotal:    load.MemTotal,
					MemUsed:     load.MemUsed,
					DiskTotal:   load.DiskTotal,
					DiskUsed:    load.DiskUsed,
					NetInSpeed:  load.NetInSpeed,
					NetOutSpeed: load.NetOutSpeed,
				})
		}
	}
}

// startNode starts a runNode goroutine for the given node ID.
func (o *Orchestrator) startNode(ctx context.Context, nodeID int) {
	o.mu.Lock()
	if _, exists := o.nodes[nodeID]; exists {
		o.mu.Unlock()
		return
	}
	nodeCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	o.nodes[nodeID] = &nodeHandle{cancel: cancel, done: done}
	o.mu.Unlock()

	log.Printf("[machine] Starting node %d", nodeID)

	go func() {
		defer close(done)
		o.runNode(nodeCtx, nodeID)
	}()
}

// runNode manages the full lifecycle of a single node in machine mode.
func (o *Orchestrator) runNode(ctx context.Context, nodeID int) {
	prefix := fmt.Sprintf("[machine:node%d] ", nodeID)
	cfg := o.nodeCfg[nodeID]

	// Create per-node client
	nodeClient := o.client.ForNode(nodeID)
	sbManager := proxy.New(cfg.Singbox)
	statsCol := collector.New(cfg.Singbox.StatsURL, fmt.Sprintf("%d", nodeID))
	deviceLimitEnforcer := devicelimit.New(cfg.Singbox.StatsURL)

	// Pull initial config
	nodeConfig, users, err := pullNodeConfig(nodeClient, prefix)
	if err != nil {
		log.Printf("%sFailed to get initial config: %v", prefix, err)
		return
	}
	if nodeConfig == nil {
		log.Printf("%sPanel returned empty config", prefix)
		return
	}

	// Start sing-box
	if err := startNodeSingbox(sbManager, nodeConfig, cfg, users, prefix); err != nil {
		log.Printf("%sFailed to start sing-box: %v", prefix, err)
		return
	}
	applyKnownUsers(statsCol, deviceLimitEnforcer, users)

	// Start crash recovery watcher
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	go watchNodeSingbox(watchCtx, sbManager, nodeClient, cfg, prefix)

	// Initial device limit sync
	if limits, err := nodeClient.FetchDeviceLimit(); err == nil {
		deviceLimitEnforcer.UpdateLimits(limits)
		if len(limits) > 0 {
			log.Printf("%sDevice limits synced: %d users with limits", prefix, len(limits))
		}
	}

	// Tickers
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()
	statsTicker := time.NewTicker(statsInterval)
	defer statsTicker.Stop()
	aliveTicker := time.NewTicker(aliveInterval)
	defer aliveTicker.Stop()
	deviceLimitTicker := time.NewTicker(deviceLimitCheckInterval)
	defer deviceLimitTicker.Stop()
	syncLimitsTicker := time.NewTicker(deviceLimitSyncInterval)
	defer syncLimitsTicker.Stop()

	startTime := time.Now()
	configFailures := 0
	pendingTraffic := make(map[string][2]int64)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%sStopping...", prefix)
			if sbManager.IsRunning() {
				sbManager.Stop()
			}
			return

		case <-heartbeatTicker.C:
			if !sbManager.IsRunning() {
				continue
			}
			cpu, mem := getSystemStats()
			uptime := uint64(time.Since(startTime).Seconds())
			configChanged, _, err := nodeClient.Heartbeat(cpu, mem, uptime)
			if err != nil {
				log.Printf("%sHeartbeat error: %v", prefix, err)
				continue
			}
			if configChanged {
				log.Printf("%sConfig change detected, applying...", prefix)
				newNodeConfig, newUsers, err := pullNodeConfig(nodeClient, prefix)
				if err != nil {
					configFailures++
					log.Printf("%sConfig pull failed (%d/%d): %v", prefix, configFailures, maxConfigFailures, err)
					if configFailures >= maxConfigFailures {
						log.Printf("%sToo many failures, restarting sing-box...", prefix)
						if sbManager.IsRunning() {
							sbManager.Stop()
						}
						configFailures = 0
					}
					continue
				}
				configFailures = 0
				if newNodeConfig == nil {
					log.Printf("%sNew config is empty, stopping sing-box", prefix)
					sbManager.Stop()
				} else if err := hotReloadOrRestartNode(sbManager, newNodeConfig, cfg, newUsers, prefix); err != nil {
					log.Printf("%sFailed to apply config: %v", prefix, err)
				} else {
					applyKnownUsers(statsCol, deviceLimitEnforcer, newUsers)
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
// Merge with pending buffer
				for uuid, delta := range trafficData {
					pendingTraffic[uuid] = [2]int64{
						pendingTraffic[uuid][0] + delta[0],
						pendingTraffic[uuid][1] + delta[1],
					}
				}
				// Cap pending buffer size to prevent memory leak
				const maxPendingUsers = 5000
				if len(pendingTraffic) > maxPendingUsers {
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
				aliveIPs = map[string][]string{}
			}
			cpu, mem := getSystemStats()
			if err := nodeClient.Report(pendingTraffic, aliveIPs, cpu, mem, 0); err != nil {
				log.Printf("%sReport failed (will retry): %v", prefix, err)
			} else {
				pendingTraffic = make(map[string][2]int64)
			}

		case <-aliveTicker.C:
			if !sbManager.IsRunning() {
				continue
			}
			aliveIPs, err := statsCol.CollectAliveIPs()
			if err != nil {
				continue
			}
			nodeClient.ReportAlive(aliveIPs)

		case <-deviceLimitTicker.C:
			if !sbManager.IsRunning() || !deviceLimitEnforcer.HasLimits() {
				continue
			}
			closed, err := deviceLimitEnforcer.Enforce()
			if err != nil {
				log.Printf("%sDevice limit error: %v", prefix, err)
				continue
			}
			if closed > 0 {
				log.Printf("%sDevice limit: closed %d excess connections", prefix, closed)
			}

		case <-syncLimitsTicker.C:
			limits, err := nodeClient.FetchDeviceLimit()
			if err != nil {
				continue
			}
			deviceLimitEnforcer.UpdateLimits(limits)
		}
	}
}

func (o *Orchestrator) stopNode(nodeID int) {
	o.mu.Lock()
	h, ok := o.nodes[nodeID]
	if !ok {
		o.mu.Unlock()
		return
	}
	delete(o.nodes, nodeID)
	o.mu.Unlock()

	log.Printf("[machine] Stopping node %d", nodeID)
	h.cancel()
	<-h.done
}

func (o *Orchestrator) stopAll() {
	o.mu.Lock()
	handles := make(map[int]*nodeHandle)
	for id, h := range o.nodes {
		handles[id] = h
	}
	o.mu.Unlock()

	for id, h := range handles {
		log.Printf("[machine] Stopping node %d", id)
		h.cancel()
	}
	for _, h := range handles {
		<-h.done
	}
}

func (o *Orchestrator) rediscover(ctx context.Context) {
	nodesResp, err := o.client.GetMachineNodes()
	if err != nil {
		log.Printf("[machine] Node discovery failed: %v", err)
		return
	}

	wanted := make(map[int]bool, len(nodesResp.Nodes))
	for _, n := range nodesResp.Nodes {
		wanted[int(n.ID)] = true
	}

	o.mu.Lock()
	var toRemove []int
	for id := range o.nodes {
		if !wanted[id] {
			toRemove = append(toRemove, id)
		}
	}
	o.mu.Unlock()

	for _, id := range toRemove {
		o.stopNode(id)
	}

	for _, n := range nodesResp.Nodes {
		o.startNode(ctx, int(n.ID))
	}
}

func (o *Orchestrator) connectWS(ctx context.Context) {
	hs, err := o.client.Handshake()
	if err != nil {
		log.Printf("[machine] Handshake failed: %v", err)
		return
	}
	if !hs.WebSocket.Enabled || hs.WebSocket.WSURL == "" {
		log.Printf("[machine] WS not enabled by panel, using HTTP polling")
		return
	}

	wsClient := wsclient.NewMachineClient(hs.WebSocket.WSURL, o.cfg.Machine.Token, o.cfg.Machine.ID)

	wsClient.RegisterHandler("restart", func(cmd wsclient.Command) error {
		nodeID := extractNodeID(cmd)
		if nodeID > 0 {
			log.Printf("[machine] Restarting node %d", nodeID)
			o.stopNode(nodeID)
			o.startNode(context.Background(), nodeID)
		}
		return nil
	})

	wsClient.RegisterHandler("reload", func(cmd wsclient.Command) error {
		nodeID := extractNodeID(cmd)
		if nodeID <= 0 {
			return nil
		}
		prefix := fmt.Sprintf("[machine:node%d] ", nodeID)
		log.Printf("%sReceived reload command, restarting node...", prefix)
		o.stopNode(nodeID)
		o.startNode(context.Background(), nodeID)
		return nil
	})

wsClient.RegisterHandler("sync.nodes", func(cmd wsclient.Command) error {
			log.Printf("[machine] sync.nodes received, rediscovering...")
			go o.rediscover(context.Background())
			return nil
		})

		// update: panel pushes a new binary URL — download, replace
		wsClient.RegisterHandler("update", func(cmd wsclient.Command) error {
			log.Printf("[machine] Received update command")
			var data struct {
				DownloadURL string `json:"download_url"`
			}
			if err := json.Unmarshal(cmd.Data, &data); err != nil {
				return fmt.Errorf("parse update command: %w", err)
			}
			if data.DownloadURL == "" {
				return fmt.Errorf("update command missing download_url")
			}

			// Determine binary path from the first running node's config
			binaryPath := o.cfg.Singbox.BinaryPath
			if binaryPath == "" {
				binaryPath = "sing-box"
			}

			tmpPath := binaryPath + ".tmp"
			if err := downloadFile(data.DownloadURL, tmpPath); err != nil {
				return fmt.Errorf("download update: %w", err)
			}
			if err := os.Chmod(tmpPath, 0755); err != nil {
				os.Remove(tmpPath)
				return fmt.Errorf("chmod update: %w", err)
			}

			// Stop all nodes gracefully
			o.stopAll()

			// Replace binary
			if err := os.Rename(tmpPath, binaryPath); err != nil {
				os.Remove(tmpPath)
				return fmt.Errorf("replace binary: %w", err)
			}

			log.Printf("[machine] Binary updated, nodes will use new binary on next start")
			return nil
		})

// On reconnect, trigger a full sync for all running nodes
		wsClient.OnReconnect = func() {
			log.Printf("[machine] WS reconnected, syncing all running nodes...")
			// First rediscover to get fresh node list
			o.rediscover(context.Background())

			// Then pull fresh config+users for every currently running node
			o.mu.Lock()
			nodeIDs := make([]int, 0, len(o.nodes))
			for id := range o.nodes {
				nodeIDs = append(nodeIDs, id)
			}
			o.mu.Unlock()

			for _, id := range nodeIDs {
				prefix := fmt.Sprintf("[machine:node%d] ", id)
				cfg := o.nodeCfg[id]
				nodeClient := o.client.ForNode(id)
				sbManager := proxy.New(cfg.Singbox)

				nodeConfig, users, err := pullNodeConfig(nodeClient, prefix)
				if err != nil {
					log.Printf("%sPost-reconnect config sync failed: %v", prefix, err)
					continue
				}
				if nodeConfig == nil {
					continue
				}
				if err := hotReloadOrRestartNode(sbManager, nodeConfig, cfg, users, prefix); err != nil {
					log.Printf("%sPost-reconnect reload failed: %v", prefix, err)
				}
			}
		}

	if err := wsClient.ConnectMachine(); err != nil {
		log.Printf("[machine] WS connection failed: %v", err)
		return
	}

	go wsClient.ReconnectLoop()
	o.wsClient = wsClient
	log.Printf("[machine] WS connected")
}

// extractNodeID extracts node_id from command data payload.
func extractNodeID(cmd wsclient.Command) int {
	if cmd.Data == nil {
		return 0
	}
	var data map[string]interface{}
	if err := json.Unmarshal(cmd.Data, &data); err != nil {
		return 0
	}
	if id, ok := data["node_id"].(float64); ok {
		return int(id)
	}
	return 0
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

// pullNodeConfig fetches node config and users, converting to kernel types.
func pullNodeConfig(client *httpclient.Client, prefix string) (*kernel.NodeConfigFromPanel, []kernel.User, error) {
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
		return nil, nil, fmt.Errorf("config pull timed out")
	case r := <-ch:
		return r.config, r.users, r.err
	}
}

// startNodeSingbox starts sing-box with the given config.
func startNodeSingbox(sbManager *proxy.SingboxManager, nodeConfig *kernel.NodeConfigFromPanel, cfg config.NormalizedNode, users []kernel.User, prefix string) error {
	if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
		log.Printf("%sStarting sing-box (json mode)", prefix)
		return sbManager.Start(nodeConfig.ConfigJSON)
	}

	kernelCfg := nodeConfig.ToNodeConfig()
	kernelCfg.StatsPort = cfg.Singbox.StatsPort

	configJSON, err := kernel.GenerateSingboxConfig(kernelCfg, users)
	if err != nil {
		return fmt.Errorf("generate config: %w", err)
	}
	log.Printf("%sStarting sing-box (protocol=%s, port=%d, %d users)", prefix, nodeConfig.Protocol, nodeConfig.ServerPort, len(users))
	return sbManager.Start(configJSON)
}

// hotReloadOrRestartNode tries hot reload, falls back to full restart.
func hotReloadOrRestartNode(sbManager *proxy.SingboxManager, nodeConfig *kernel.NodeConfigFromPanel, cfg config.NormalizedNode, users []kernel.User, prefix string) error {
	var configJSON string
	if nodeConfig.ConfigMode == "json" && nodeConfig.ConfigJSON != "" {
		configJSON = nodeConfig.ConfigJSON
	} else {
		kernelCfg := nodeConfig.ToNodeConfig()
		kernelCfg.StatsPort = cfg.Singbox.StatsPort
		var err error
		configJSON, err = kernel.GenerateSingboxConfig(kernelCfg, users)
		if err != nil {
			return fmt.Errorf("generate config: %w", err)
		}
	}

	if sbManager.IsRunning() {
		if err := sbManager.ReloadConfig(configJSON); err != nil {
			log.Printf("%sHot reload failed (%v), falling back to restart", prefix, err)
			return sbManager.Restart(configJSON)
		}
		log.Printf("%sConfig hot-reloaded (protocol=%s, %d users)", prefix, nodeConfig.Protocol, len(users))
		return nil
	}

	return sbManager.Start(configJSON)
}

// watchNodeSingbox monitors sing-box and restarts with backoff on crash.
func watchNodeSingbox(ctx context.Context, sbManager *proxy.SingboxManager, client *httpclient.Client, cfg config.NormalizedNode, prefix string) {
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

				nodeConfig, users, err := pullNodeConfig(client, prefix)
				if err != nil {
					log.Printf("%sFailed to pull config for restart: %v", prefix, err)
					wasRunning = false
					continue
				}
				if nodeConfig == nil {
					wasRunning = false
					continue
				}

				if err := startNodeSingbox(sbManager, nodeConfig, cfg, users, prefix); err != nil {
					log.Printf("%sFailed to restart sing-box: %v", prefix, err)
				} else {
					log.Printf("%ssing-box restarted by watcher (attempt #%d)", prefix, consecutiveFailures)
				}
			}

			if isRunning && consecutiveFailures > 0 {
				consecutiveFailures = 0
			}

			wasRunning = isRunning
		}
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

	// downloadFile downloads a URL to a local path.
	func downloadFile(url, path string) error {
		log.Printf("[machine] Downloading %s -> %s", url, path)

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

		log.Printf("[machine] Downloaded %d bytes", written)
		return nil
	}
