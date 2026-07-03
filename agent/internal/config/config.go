package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds the full agent configuration.
// Values can come from YAML file, environment variables, or CLI flags.
// Priority: CLI flags > environment variables > YAML file.
type Config struct {
	Panel PanelConfig `yaml:"panel"`
	Nodes []NodeConfig `yaml:"nodes"`
}

// PanelConfig defines how the agent connects to the Nexus panel.
type PanelConfig struct {
	Address string `yaml:"address"`
}

// NodeConfig defines a single proxy node.
type NodeConfig struct {
	Name string `yaml:"name"`
	Token string `yaml:"token"`
	Address string `yaml:"address"`
	Singbox SingboxConfig `yaml:"singbox"`
}

// SingboxConfig defines where to find and how to run sing-box.
type SingboxConfig struct {
	BinaryPath string `yaml:"binary_path"`
	ConfigPath string `yaml:"config_path"`
	WorkingDir string `yaml:"working_dir"`
	StatsURL string `yaml:"stats_url"`
	StatsPort int `yaml:"stats_port"`
}

// CLIArgs holds command-line arguments for quick setup.
// Usage: ./agent --panel https://panel.com --token REGISTER_TOKEN --name my-node
type CLIArgs struct {
	PanelURL   string
	Token      string
	Name       string
	ConfigPath string
}

// ParseCLIArgs parses environment variables and CLI flags.
// Supports both --flag value and ENV_VAR=value formats.
//
// Environment variables:
//   NEXUS_PANEL_URL  - panel HTTP address
//   NEXUS_TOKEN      - node registration token
//   NEXUS_NODE_NAME  - node display name
//   NEXUS_CONFIG     - path to config file
//
// CLI flags:
//   --panel  URL
//   --token  TOKEN
//   --name   NODE_NAME
//   --config PATH
func ParseCLIArgs(args []string) CLIArgs {
	var cli CLIArgs
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--panel":
			if i+1 < len(args) { i++; cli.PanelURL = args[i] }
		case "--token":
			if i+1 < len(args) { i++; cli.Token = args[i] }
		case "--name":
			if i+1 < len(args) { i++; cli.Name = args[i] }
		case "--config":
			if i+1 < len(args) { i++; cli.ConfigPath = args[i] }
		}
	}

	// Environment variables override CLI defaults
	if v := os.Getenv("NEXUS_PANEL_URL"); v != "" { cli.PanelURL = v }
	if v := os.Getenv("NEXUS_TOKEN"); v != "" { cli.Token = v }
	if v := os.Getenv("NEXUS_NODE_NAME"); v != "" { cli.Name = v }
	if v := os.Getenv("NEXUS_CONFIG"); v != "" { cli.ConfigPath = v }

	if cli.ConfigPath == "" {
		cli.ConfigPath = "agent.yaml"
	}

	return cli
}

// LoadWithCLI loads config from YAML and applies CLI/env overrides.
// If --panel and --token are provided, a single-node config is auto-generated
// without requiring a YAML file.
func LoadWithCLI(args []string) (Config, error) {
	cli := ParseCLIArgs(args)

	// Quick mode: --panel + --token provided, no YAML needed
	if cli.PanelURL != "" && cli.Token != "" {
		name := cli.Name
		if name == "" { name = "node-1" }
		cfg := Config{
			Panel: PanelConfig{Address: cli.PanelURL},
			Nodes: []NodeConfig{{
				Name:  name,
				Token: cli.Token,
			}}}
		if err := cfg.Validate(); err != nil {
			return cfg, fmt.Errorf("validate quick config: %w", err)
		}
		return cfg, nil
	}

	// YAML mode: load from file
	cfg, err := Load(cli.ConfigPath)
	if err != nil {
		return cfg, err
	}

	// Apply CLI overrides
	if cli.PanelURL != "" { cfg.Panel.Address = cli.PanelURL }
	if len(cfg.Nodes) > 0 && cli.Token != "" {
		cfg.Nodes[0].Token = cli.Token
	}
	if len(cfg.Nodes) > 0 && cli.Name != "" {
		cfg.Nodes[0].Name = cli.Name
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Panel.Address == "" {
		return fmt.Errorf("panel.address is required (use --panel or NEXUS_PANEL_URL)")
	}
	if len(c.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured (use --token or NEXUS_TOKEN)")
	}
	for i, n := range c.Nodes {
		if n.Token == "" {
			return fmt.Errorf("nodes[%d].token is required", i)
		}
		if n.Singbox.BinaryPath == "" {
			c.Nodes[i].Singbox.BinaryPath = "sing-box"
		}
		if n.Singbox.ConfigPath == "" {
			c.Nodes[i].Singbox.ConfigPath = fmt.Sprintf("singbox-%d.json", i)
		}
		if n.Singbox.WorkingDir == "" {
			c.Nodes[i].Singbox.WorkingDir = "."
		}
		if n.Singbox.StatsURL == "" {
			port := n.Singbox.StatsPort
			if port == 0 { port = 9090 }
			c.Nodes[i].Singbox.StatsURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		}
	}
	return nil
}

// Load reads config from a YAML file.
func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// StatsPortInt is a helper to parse stats port from string.
func StatsPortInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil { return def }
	return n
}
