package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full agent configuration.
type Config struct {
	Panel PanelConfig `yaml:"panel"`
	Nodes []NodeConfig `yaml:"nodes"`
}

// PanelConfig defines how the agent connects to the Nexus panel.
type PanelConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`  // global server_token from panel settings
}

// NodeConfig defines a single proxy node.
type NodeConfig struct {
	NodeID  int           `yaml:"node_id"`
	Address string        `yaml:"address"`
	Singbox SingboxConfig `yaml:"singbox"`
}

// SingboxConfig defines where to find and how to run sing-box.
type SingboxConfig struct {
	BinaryPath string `yaml:"binary_path"`
	ConfigPath string `yaml:"config_path"`
	WorkingDir string `yaml:"working_dir"`
	StatsURL   string `yaml:"stats_url"`
	StatsPort  int    `yaml:"stats_port"`
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

// Save writes the configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Panel.URL == "" {
		return fmt.Errorf("panel.url is required")
	}
	if c.Panel.Token == "" {
		return fmt.Errorf("panel.token (server_token) is required")
	}
	if len(c.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured")
	}
	for i := range c.Nodes {
		n := &c.Nodes[i]
		if n.NodeID == 0 {
			return fmt.Errorf("nodes[%d].node_id is required", i)
		}
		if n.Singbox.BinaryPath == "" {
			n.Singbox.BinaryPath = "sing-box"
		}
		if n.Singbox.ConfigPath == "" {
			n.Singbox.ConfigPath = fmt.Sprintf("singbox-%d.json", n.NodeID)
		}
		if n.Singbox.WorkingDir == "" {
			n.Singbox.WorkingDir = "."
		}
		if n.Singbox.StatsURL == "" {
			port := n.Singbox.StatsPort
			if port == 0 {
				port = 9090
			}
			n.Singbox.StatsURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		}
	}
	return nil
}
