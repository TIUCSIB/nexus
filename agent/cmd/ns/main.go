package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"nexus-agent/internal/config"
)

const defaultConfigPath = "agent.yaml"
const defaultBinName = "agent.exe"
const serviceName = "nexus-agent"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "bind":
		cmdBind(args)
	case "list":
		cmdList(args)
	case "status":
		cmdStatus(args)
	case "service":
		cmdService(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Nexus Agent CLI (ns) - Manage proxy nodes

Usage:
  ns bind --panel <url> --token <token> --node-id <id> 
          [--address <addr>] [--stats-port <port>] [--config <path>]
    Bind a node to the panel. Creates or updates agent.yaml.

  ns list [--config <path>]
    List all configured nodes.

  ns status [--config <path>]
    Show whether the agent is running.

  ns service install [--config <path>] [--bin <path>]
    Install the agent as a system service.

  ns service restart
    Restart the agent service.

  ns service stop
    Stop the agent service.

  ns service uninstall
    Remove the agent service.

  ns help
    Show this help message.`)
}

// ──── bind ──────────────────────────────────────────────────────────────

func cmdBind(args []string) {
	fs := flag.NewFlagSet("bind", flag.ExitOnError)
	panelURL := fs.String("panel", "", "Panel URL (required)")
	token := fs.String("token", "", "Server token (required)")
	nodeID := fs.Int("node-id", 0, "Node ID (required)")
	address := fs.String("address", "0.0.0.0", "Node listen address")
	statsPort := fs.Int("stats-port", 9090, "Sing-box stats API port")
	configPath := fs.String("config", defaultConfigPath, "Config file path")
	fs.Parse(args)

	if *panelURL == "" || *token == "" || *nodeID == 0 {
		fs.Usage()
		os.Exit(1)
	}

	// Load existing config if present
	var cfg *config.Config
	if _, err := os.Stat(*configPath); err == nil {
		existing, err := config.Load(*configPath)
		if err == nil {
			cfg = &existing
		}
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Set panel config
	cfg.Panel.URL = strings.TrimRight(*panelURL, "/")
	cfg.Panel.Token = *token

	// Update or add the node
	found := false
	for i, n := range cfg.Nodes {
		if n.NodeID == *nodeID {
			cfg.Nodes[i].Address = *address
			cfg.Nodes[i].Singbox.StatsPort = *statsPort
			if cfg.Nodes[i].Singbox.ConfigPath == "" {
				cfg.Nodes[i].Singbox.ConfigPath = fmt.Sprintf("singbox-%d.json", *nodeID)
			}
			if cfg.Nodes[i].Singbox.BinaryPath == "" {
				cfg.Nodes[i].Singbox.BinaryPath = "sing-box"
			}
			if cfg.Nodes[i].Singbox.WorkingDir == "" {
				cfg.Nodes[i].Singbox.WorkingDir = "."
			}
			found = true
			fmt.Printf("Updated node %d\n", *nodeID)
			break
		}
	}
	if !found {
		cfg.Nodes = append(cfg.Nodes, config.NodeConfig{
			NodeID:  *nodeID,
			Address: *address,
			Singbox: config.SingboxConfig{
				BinaryPath: "sing-box",
				ConfigPath: fmt.Sprintf("singbox-%d.json", *nodeID),
				WorkingDir: ".",
				StatsPort:  *statsPort,
			},
		})
		fmt.Printf("Added node %d\n", *nodeID)
	}

	if err := cfg.Save(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config written to %s\n", *configPath)
}

// ──── list ──────────────────────────────────────────────────────────────

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "Config file path")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Node ID\tAddress\tStats Port\tConfig File")
	fmt.Fprintln(w, "-------\t-------\t----------\t-----------")
	for _, n := range cfg.Nodes {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", n.NodeID, n.Address, n.Singbox.StatsPort, n.Singbox.ConfigPath)
	}
	w.Flush()
}

// ──── status ────────────────────────────────────────────────────────────

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "Config file path")
	fs.Parse(args)

	// Try to load config for display
	cfg, loadErr := config.Load(*configPath)

	running := isAgentRunning()

	if running {
		fmt.Println("Status: RUNNING")
	} else {
		fmt.Println("Status: STOPPED")
	}

	if loadErr == nil {
		fmt.Printf("Panel: %s\n", cfg.Panel.URL)
		fmt.Printf("Nodes: %d configured\n", len(cfg.Nodes))
	}
}

func isAgentRunning() bool {
	// On Windows use tasklist, on Linux/macOS use pgrep
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq agent.exe", "/NH")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Count(string(out), "agent.exe") > 0
	} else {
		cmd := exec.Command("pgrep", "-x", "agent")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(out)) != ""
	}
}

// ──── service ───────────────────────────────────────────────────────────

func cmdService(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ns service <install|restart|stop|uninstall>")
		os.Exit(1)
	}

	action := args[0]
	restArgs := args[1:]

	switch action {
	case "install":
		serviceInstall(restArgs)
	case "restart":
		serviceRestart()
	case "stop":
		serviceStop()
	case "uninstall":
		serviceUninstall()
	default:
		fmt.Fprintf(os.Stderr, "Unknown service action: %s\n", action)
		os.Exit(1)
	}
}

func serviceInstall(args []string) {
	fs := flag.NewFlagSet("service install", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "Config file path")
	binPath := fs.String("bin", "", "Path to the agent binary (default: auto-detect)")
	fs.Parse(args)

	// Resolve binary path
	bin := *binPath
	if bin == "" {
		exe, err := os.Executable()
		if err == nil {
			// ns binary is in same directory as agent
			dir := filepath.Dir(exe)
			bin = filepath.Join(dir, defaultBinName)
		} else {
			bin = defaultBinName
		}
	}
	absBin, _ := filepath.Abs(bin)

	absConfig, _ := filepath.Abs(*configPath)

	if runtime.GOOS == "windows" {
		serviceInstallWindows(absBin, absConfig)
	} else {
		serviceInstallLinux(absBin, absConfig)
	}
}

func serviceInstallWindows(binPath, configPath string) {
	// Use sc.exe to create the service
	cmd := exec.Command("sc", "create", serviceName,
		"binPath=", fmt.Sprintf(`"%s" -config "%s"`, binPath, configPath),
		"start=", "auto",
		"displayName=", "Nexus Agent",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create service: %v\n%s\n", err, string(output))
		os.Exit(1)
	}
	fmt.Printf("Service '%s' created\n", serviceName)

	// Start the service
	start := exec.Command("sc", "start", serviceName)
	if startOut, startErr := start.CombinedOutput(); startErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: service created but failed to start: %v\n%s\n", startErr, string(startOut))
	} else {
		fmt.Println("Service started")
	}
}

func serviceInstallLinux(binPath, configPath string) {
	unitDir := "/etc/systemd/system"
	unitPath := filepath.Join(unitDir, serviceName+".service")

	unitContent := fmt.Sprintf(`[Unit]
Description=Nexus Agent
After=network.target

[Service]
Type=simple
ExecStart=%s -config %s
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
`, binPath, configPath)

	if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write systemd unit: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Systemd unit created: %s\n", unitPath)

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", serviceName).Run()
	exec.Command("systemctl", "start", serviceName).Run()
	fmt.Println("Service enabled and started")
}

func serviceRestart() {
	if runtime.GOOS == "windows" {
		exec.Command("sc", "stop", serviceName).Run()
		exec.Command("sc", "start", serviceName).Run()
	} else {
		exec.Command("systemctl", "restart", serviceName).Run()
	}
	fmt.Println("Service restarted")
}

func serviceStop() {
	if runtime.GOOS == "windows" {
		exec.Command("sc", "stop", serviceName).Run()
	} else {
		exec.Command("systemctl", "stop", serviceName).Run()
	}
	fmt.Println("Service stopped")
}

func serviceUninstall() {
	if runtime.GOOS == "windows" {
		exec.Command("sc", "stop", serviceName).Run()
		exec.Command("sc", "delete", serviceName).Run()
	} else {
		exec.Command("systemctl", "stop", serviceName).Run()
		exec.Command("systemctl", "disable", serviceName).Run()
		os.Remove("/etc/systemd/system/" + serviceName + ".service")
		exec.Command("systemctl", "daemon-reload").Run()
	}
	fmt.Println("Service uninstalled")
}