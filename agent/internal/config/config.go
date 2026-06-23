package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full agent configuration loaded from YAML.
type Config struct {
	Panel PanelConfig `yaml:"panel"`
	Nodes []NodeConfig `yaml:"nodes"`
}

// PanelConfig defines how the agent connects to the Nexus panel.
type PanelConfig struct {
	Address string `yaml:"address"` // panel HTTP address, e.g. "http://panel.example.com:8080"
}

// NodeConfig defines a single proxy node managed by this agent.
type NodeConfig struct {
	Name    string         `yaml:"name"`
	Token   string         `yaml:"token"`
	Address string         `yaml:"address"`
	Singbox SingboxConfig  `yaml:"singbox"`
}

// SingboxConfig defines where to find and how to run sing-box.
type SingboxConfig struct {
	BinaryPath string `yaml:"binary_path"` // path to sing-box binary, default "sing-box"
	ConfigPath string `yaml:"config_path"` // where to write sing-box config, default "singbox.json"
	WorkingDir string `yaml:"working_dir"` // working directory for sing-box
	StatsURL   string `yaml:"stats_url"`   // sing-box stats API, default "http://127.0.0.1:9090"
	StatsPort  int    `yaml:"stats_port"`   // sing-box stats port, used to build StatsURL if empty
}

// Validate checks that required fields are present and fills defaults.
func (c *Config) Validate() error {
	if c.Panel.Address == "" {
		return fmt.Errorf("panel.address is required")
	}
	if len(c.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured under 'nodes'")
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
			if port == 0 {
				port = 9090
			}
			c.Nodes[i].Singbox.StatsURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		}
	}
	return nil
}

// Load reads the agent configuration from a YAML file.
func Load(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
