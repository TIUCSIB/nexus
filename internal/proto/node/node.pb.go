// Package node defines the gRPC message types for the Nexus NodeService.
// Hand-written to mirror protoc output - compiles without protoc.
package node

import "encoding/json"

// RegisterRequest is sent by a node agent to register with the panel.
type RegisterRequest struct {
	NodeName string `json:"node_name"`
	Address  string `json:"address"`
	Token    string `json:"token"`
}

func (m *RegisterRequest) Reset()                { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *RegisterRequest) ProtoMessage()          {}

// RegisterResponse is returned after a successful node registration.
type RegisterResponse struct {
	Success   bool   `json:"success"`
	NodeID    string `json:"node_id"`
	AuthToken string `json:"auth_token"`
	Error     string `json:"error,omitempty"`
}

func (m *RegisterResponse) Reset()                { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *RegisterResponse) ProtoMessage()          {}

// HeartbeatRequest is sent periodically by a node agent.
type HeartbeatRequest struct {
	NodeID      string  `json:"node_id"`
	CpuUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Uptime      uint64  `json:"uptime"`
}

func (m *HeartbeatRequest) Reset()                { *m = HeartbeatRequest{} }
func (m *HeartbeatRequest) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *HeartbeatRequest) ProtoMessage()          {}

// HeartbeatResponse tells the node whether its config has changed.
type HeartbeatResponse struct {
	Success       bool `json:"success"`
	ConfigChanged bool `json:"config_changed"`
}

func (m *HeartbeatResponse) Reset()                { *m = HeartbeatResponse{} }
func (m *HeartbeatResponse) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *HeartbeatResponse) ProtoMessage()          {}

// GetConfigRequest asks the panel for the current sing-box configuration.
type GetConfigRequest struct {
	NodeID string `json:"node_id"`
}

func (m *GetConfigRequest) Reset()                { *m = GetConfigRequest{} }
func (m *GetConfigRequest) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *GetConfigRequest) ProtoMessage()          {}

// GetConfigResponse carries the generated sing-box config and user list.
type GetConfigResponse struct {
	SingboxConfig string `json:"singbox_config"`
	UsersJSON     string `json:"users_json"`
}

func (m *GetConfigResponse) Reset()                { *m = GetConfigResponse{} }
func (m *GetConfigResponse) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *GetConfigResponse) ProtoMessage()          {}

// TrafficReport contains per-user traffic counters collected by a node.
type TrafficReport struct {
	NodeID  string          `json:"node_id"`
	Entries []*TrafficEntry `json:"entries"`
}

func (m *TrafficReport) Reset()                { *m = TrafficReport{} }
func (m *TrafficReport) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *TrafficReport) ProtoMessage()          {}

// TrafficEntry holds upload/download bytes for a single user on a node.
type TrafficEntry struct {
	UserUUID string `json:"user_uuid"`
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

func (m *TrafficEntry) Reset()                { *m = TrafficEntry{} }
func (m *TrafficEntry) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *TrafficEntry) ProtoMessage()          {}

// TrafficReportResponse acknowledges a traffic report.
type TrafficReportResponse struct {
	Success bool `json:"success"`
}

func (m *TrafficReportResponse) Reset()                { *m = TrafficReportResponse{} }
func (m *TrafficReportResponse) String() string         { b, _ := json.Marshal(m); return string(b) }
func (m *TrafficReportResponse) ProtoMessage()          {}
