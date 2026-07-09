package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full agent configuration.
// Compatible with Xboard-Node configuration format.
type Config struct {
	// Top-level panel config (for single node mode)
	Panel PanelConfig `yaml:"panel"`

	// Top-level machine config (for machine mode)
	Machine *MachineConfig `yaml:"machine,omitempty"`

	// Top-level singbox config (inherited by all nodes)
	Singbox SingboxConfig `yaml:"singbox"`

	// Top-level kernel config (for Xboard compatibility)
	Kernel KernelConfig `yaml:"kernel"`

	// Top-level log config (for Xboard compatibility)
	Log LogConfig `yaml:"log"`

	// Health check HTTP server port (0 = disabled)
	HealthPort int `yaml:"health_port,omitempty"`

	// Multi-node mode: list of instances
	Instances []InstanceConfig `yaml:"instances"`

	// Legacy Nexus format: node list
	Nodes []NodeConfig `yaml:"nodes"`
}

// MachineConfig identifies this process as a panel-managed machine that
// dynamically discovers and runs all nodes bound to it.
type MachineConfig struct {
	ID    int    `yaml:"id"`
	Token string `yaml:"token"`
	PanelURL string `yaml:"panel_url,omitempty"`
}

// IsMachineMode returns true if the config is set for machine mode.
func (c *Config) IsMachineMode() bool {
	return c.Machine != nil && c.Machine.ID > 0 && c.Machine.Token != ""
}

// PanelConfig defines how the agent connects to the panel.
type PanelConfig struct {
	URL    string `yaml:"url"`
	Token  string `yaml:"token"`
	NodeID int    `yaml:"node_id"`
}

// InstanceConfig defines a single node instance (Xboard format).
type InstanceConfig struct {
	Panel   PanelConfig           `yaml:"panel"`
	Machine InstanceMachineConfig `yaml:"machine,omitempty"`
	Singbox SingboxConfig         `yaml:"singbox,omitempty"`
	Log     LogConfig             `yaml:"log,omitempty"`
}

// InstanceMachineConfig defines machine-level authentication (Xboard format).
// Used within InstanceConfig for per-instance machine auth (legacy).
type InstanceMachineConfig struct {
	MachineID int    `yaml:"machine_id"`
	Token     string `yaml:"token"`
}

// KernelConfig defines kernel settings (Xboard format).
type KernelConfig struct {
	Type string `yaml:"type"`
}

// LogConfig defines logging settings.
type LogConfig struct {
	Level    string `yaml:"level"`
	LogFile  string `yaml:"log_file,omitempty"`
	LogMaxSize int64 `yaml:"log_max_size,omitempty"`
}

// NodeConfig defines a single proxy node (Nexus format).
type NodeConfig struct {
	NodeID  string        `yaml:"node_id"`
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

// NormalizedNode is the internal representation after normalization.
type NormalizedNode struct {
	NodeID     int
	Token      string
	MachineID  int
	PanelURL   string
	PanelToken string
	Singbox    SingboxConfig
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

// NormalizeNodes converts the config into a list of NormalizedNode.
// Supports both Xboard format (instances/panel.node_id) and Nexus format (nodes[]).
func (c *Config) NormalizeNodes() ([]NormalizedNode, error) {
	var nodes []NormalizedNode

	// Xboard format: instances[]
	if len(c.Instances) > 0 {
		for i, inst := range c.Instances {
			node := NormalizedNode{
				NodeID:     inst.Panel.NodeID,
				PanelURL:   inst.Panel.URL,
				PanelToken: inst.Panel.Token,
				MachineID:  inst.Machine.MachineID,
				Singbox:    c.mergeSingbox(inst.Singbox),
			}

			// Inherit top-level panel config
			if node.PanelURL == "" {
				node.PanelURL = c.Panel.URL
			}
			if node.PanelToken == "" && inst.Machine.Token == "" {
				node.PanelToken = c.Panel.Token
			}
			if inst.Machine.Token != "" {
				node.PanelToken = inst.Machine.Token
			}

			// Validate
			if node.NodeID == 0 && node.MachineID == 0 {
				return nil, fmt.Errorf("instances[%d]: node_id or machine_id is required", i)
			}
			if node.PanelURL == "" {
				return nil, fmt.Errorf("instances[%d]: panel.url is required", i)
			}
			if node.PanelToken == "" {
				return nil, fmt.Errorf("instances[%d]: panel.token is required", i)
			}

			nodes = append(nodes, node)
		}
		return nodes, nil
	}

	// Xboard format: single node at top level (panel.node_id)
	if c.Panel.NodeID > 0 {
			node := NormalizedNode{
				NodeID:     c.Panel.NodeID,
				PanelURL:   c.Panel.URL,
				PanelToken: c.Panel.Token,
				Singbox:    c.normalizeSingbox(c.Singbox),
			}
		if node.PanelURL == "" {
			return nil, fmt.Errorf("panel.url is required")
		}
		if node.PanelToken == "" {
			return nil, fmt.Errorf("panel.token is required")
		}
		return []NormalizedNode{node}, nil
	}

	// Nexus format: nodes[]
	if len(c.Nodes) > 0 {
		for i, n := range c.Nodes {
			nodeID := 0
			fmt.Sscanf(n.NodeID, "%d", &nodeID)

			node := NormalizedNode{
				NodeID:     nodeID,
				PanelURL:   c.Panel.URL,
				PanelToken: c.Panel.Token,
				Singbox:    c.mergeSingbox(n.Singbox),
			}

			if node.NodeID == 0 {
				return nil, fmt.Errorf("nodes[%d].node_id is required", i)
			}
			if node.PanelURL == "" {
				return nil, fmt.Errorf("panel.url is required")
			}
			if node.PanelToken == "" {
				return nil, fmt.Errorf("panel.token is required")
			}

			nodes = append(nodes, node)
		}
		return nodes, nil
	}

	return nil, fmt.Errorf("no nodes configured: use panel.node_id, instances[], or nodes[]")
}

// mergeSingbox merges node-level singbox config with top-level defaults.
func (c *Config) mergeSingbox(nodeSingbox SingboxConfig) SingboxConfig {
	result := c.Singbox // copy top-level

	if nodeSingbox.BinaryPath != "" {
		result.BinaryPath = nodeSingbox.BinaryPath
	}
	if nodeSingbox.ConfigPath != "" {
		result.ConfigPath = nodeSingbox.ConfigPath
	}
	if nodeSingbox.WorkingDir != "" {
		result.WorkingDir = nodeSingbox.WorkingDir
	}
	if nodeSingbox.StatsURL != "" {
		result.StatsURL = nodeSingbox.StatsURL
	}
	if nodeSingbox.StatsPort != 0 {
		result.StatsPort = nodeSingbox.StatsPort
	}

	return c.normalizeSingbox(result)
}

func (c *Config) normalizeSingbox(s SingboxConfig) SingboxConfig {
	if s.StatsPort == 0 {
		s.StatsPort = 9090
	}
	if s.StatsURL == "" {
		s.StatsURL = fmt.Sprintf("http://127.0.0.1:%d", s.StatsPort)
	}
	return s
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
