package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TrafficEntry represents a single user traffic record to report to the panel.
type TrafficEntry struct {
	NodeID   string `json:"node_id"`
	UserUUID string `json:"user_uuid"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// HandshakeResponse is the response from POST /api/internal/agent/handshake.
type HandshakeResponse struct {
	WebSocket WebSocketConfig `json:"websocket"`
	Settings  SettingsConfig  `json:"settings"`
}

type WebSocketConfig struct {
	Enabled bool   `json:"enabled"`
	WSURL   string `json:"ws_url,omitempty"`
}

type SettingsConfig struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

// NodeConfigResponse is the response from GET /api/internal/agent/:node_id/config.
type NodeConfigResponse struct {
	ConfigMode        string                 `json:"config_mode,omitempty"`   // "auto" or "json"
	ConfigJSON        string                 `json:"config_json,omitempty"`   // raw config for json mode
	NodeID            int                    `json:"node_id,omitempty"`
	Protocol          string                 `json:"protocol"`
	ListenIP          string                 `json:"listen_ip"`
	ServerPort        int                    `json:"server_port"`
	Network           string                 `json:"network"`
	NetworkSettings   map[string]interface{} `json:"networkSettings,omitempty"`
	BaseConfig        BaseConfig             `json:"base_config"`
	Routes            []RouteRule            `json:"routes,omitempty"`
	KernelType        string                 `json:"kernel_type,omitempty"`
	CertConfig        CertConfig             `json:"cert_config,omitempty"`
	CustomOutbounds   []CustomOutbound       `json:"custom_outbounds,omitempty"`
	TLS               int                    `json:"tls,omitempty"`
	Flow              string                 `json:"flow,omitempty"`
	TLSSettings       map[string]interface{} `json:"tls_settings,omitempty"`
	ServerName        string                 `json:"server_name,omitempty"`
	UpMbps            int                    `json:"up_mbps,omitempty"`
	DownMbps          int                    `json:"down_mbps,omitempty"`
	ObfsPassword      string                 `json:"obfs-password,omitempty"`
	CongestionControl string                 `json:"congestion_control,omitempty"`
}

type CertConfig struct {
	CertMode    string            `json:"cert_mode"`
	Domain      string            `json:"domain"`
	Email       string            `json:"email"`
	DNSProvider string            `json:"dns_provider"`
	DNSEnv      map[string]string `json:"dns_env"`
	HTTPPort    int               `json:"http_port"`
	CertFile    string            `json:"cert_file"`
	KeyFile     string            `json:"key_file"`
	CertContent string            `json:"cert_content"`
	KeyContent  string            `json:"key_content"`
	CertDir     string            `json:"cert_dir"`
}

type CustomOutbound struct {
	Tag      string                 `json:"tag"`
	Protocol string                 `json:"protocol"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	ProxyTag string                 `json:"proxy_tag,omitempty"`
}

type BaseConfig struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

type RouteRule struct {
	ID          int                    `json:"id"`
	Match       []string               `json:"match"`
	MatchRule   map[string]interface{} `json:"match_rule,omitempty"`
	Action      string                 `json:"action"`
	ActionValue string                 `json:"action_value,omitempty"`
	ActionRule  map[string]interface{} `json:"action_rule,omitempty"`
}

// UsersResponse is the response from GET /api/internal/agent/:node_id/users.
type UsersResponse struct {
	Users []UserInfo `json:"users"`
}

type UserInfo struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  int    `json:"speed_limit"`
	DeviceLimit int    `json:"device_limit"`
}

// heartbeatResponse is what the panel returns on heartbeat.
type heartbeatResponse struct {
	ConfigChanged bool `json:"config_changed"`
	PullInterval  int  `json:"pull_interval"`
}

// deviceLimitResponse is what the panel returns on GET /internal/agent/devicelimit.
type deviceLimitResponse struct {
	Limits map[string]int `json:"limits"`
}

// apiResponse is the generic wrapper returned by all panel endpoints.
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Client communicates with the Nexus panel over HTTP REST.
type Client struct {
	baseURL    string
	token      string
	nodeID     string
	machineID  int
	httpClient *http.Client
	configETag string
	usersETag  string
	configCache *NodeConfigResponse
	usersCache  []UserInfo
}

// NewClient creates a new Client pointing at the given panel address.
func NewClient(baseURL, token string, nodeID string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		nodeID:  nodeID,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SetMachineID sets the machine ID for machine-mode authentication.
func (c *Client) SetMachineID(machineID int) {
	c.machineID = machineID
}

// ForNode returns a new client bound to a specific node_id, sharing the
// same base URL, auth and HTTP transport.
func (c *Client) ForNode(nodeID int) *Client {
	return &Client{
		baseURL:    c.baseURL,
		token:      c.token,
		nodeID:     fmt.Sprintf("%d", nodeID),
		machineID:  c.machineID,
		httpClient: c.httpClient,
	}
}

// MachineNodeInfo represents a node under a machine.
type MachineNodeInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"type"`
	Address  string `json:"host"`
	Port     int    `json:"port"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// MachineNodesResponse is the response from GET /api/internal/machine/:id/nodes.
type MachineNodesResponse struct {
	Nodes        []MachineNodeInfo `json:"nodes"`
	PullInterval int               `json:"pull_interval"`
	PushInterval int               `json:"push_interval"`
}

// LoadData is the payload sent to POST /api/internal/machine/:id/load.
type LoadData struct {
	CPU         float64 `json:"cpu"`
	MemTotal    int64   `json:"mem_total"`
	MemUsed     int64   `json:"mem_used"`
	DiskTotal   int64   `json:"disk_total"`
	DiskUsed    int64   `json:"disk_used"`
	NetInSpeed  float64 `json:"net_in_speed"`
	NetOutSpeed float64 `json:"net_out_speed"`
}

// GetMachineNodes fetches all active nodes under this machine.
func (c *Client) GetMachineNodes() (*MachineNodesResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+fmt.Sprintf("/api/internal/machine/%d/nodes", c.machineID), nil)
	if err != nil {
		return nil, fmt.Errorf("create machine nodes request: %w", err)
	}
	req.Header.Set("X-Machine-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("machine nodes request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read machine nodes response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode machine nodes response: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("get machine nodes failed: %s", apiResp.Message)
	}

	var data MachineNodesResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode machine nodes data: %w", err)
	}
	return &data, nil
}

// MachineHeartbeat sends a heartbeat for the machine.
func (c *Client) MachineHeartbeat(machineID int) error {
	body, _ := json.Marshal(map[string]interface{}{})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+fmt.Sprintf("/api/internal/machine/%d/heartbeat", machineID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Machine-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// MachineReportLoad sends system load data to the panel.
// POST /api/internal/machine/:id/load
func (c *Client) MachineReportLoad(machineID int, load LoadData) error {
	body, err := json.Marshal(load)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+fmt.Sprintf("/api/internal/machine/%d/load", machineID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Machine-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Handshake performs the initial handshake with the panel.
func (c *Client) Handshake() (*HandshakeResponse, error) {
	body := map[string]interface{}{}

	var resp apiResponse
	if err := c.post("/api/internal/agent/handshake", body, &resp); err != nil {
		return nil, fmt.Errorf("handshake request: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("handshake failed: %s", resp.Message)
	}

	var data HandshakeResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode handshake response: %w", err)
	}
	return &data, nil
}

// GetConfig fetches the node configuration for this node from the panel (Xboard-style).
func (c *Client) GetConfig() (*NodeConfigResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+fmt.Sprintf("/api/internal/agent/%s/config", c.nodeID), nil)
	if err != nil {
		return nil, fmt.Errorf("create config request: %w", err)
	}
	req.Header.Set("X-Node-Token", c.token)
	if c.configETag != "" {
		req.Header.Set("If-None-Match", c.configETag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified && c.configCache != nil {
		return c.configCache, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read config response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode config response: %w (body: %s)", err, string(body))
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("get config failed: %s", apiResp.Message)
	}

	var data NodeConfigResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode config data: %w", err)
	}
	if etag := resp.Header.Get("ETag"); etag != "" {
		c.configETag = etag
	}
	c.configCache = &data

	return &data, nil
}

// GetUsers fetches the list of active users from the panel.
func (c *Client) GetUsers() ([]UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+fmt.Sprintf("/api/internal/agent/%s/users", c.nodeID), nil)
	if err != nil {
		return nil, fmt.Errorf("create users request: %w", err)
	}
	req.Header.Set("X-Node-Token", c.token)
	if c.usersETag != "" {
		req.Header.Set("If-None-Match", c.usersETag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get users request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified && c.usersCache != nil {
		return c.usersCache, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read users response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode users response: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("get users failed: %s", apiResp.Message)
	}

	var data UsersResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode users data: %w", err)
	}
	if etag := resp.Header.Get("ETag"); etag != "" {
		c.usersETag = etag
	}
	c.usersCache = data.Users

	return data.Users, nil
}

// Heartbeat sends periodic status to the panel and returns whether config changed
// and the recommended pull interval in seconds.
func (c *Client) Heartbeat(cpu, mem float64, uptime uint64) (configChanged bool, pullInterval int, err error) {
	body := map[string]interface{}{
		"cpu":    cpu,
		"mem":    mem,
		"uptime": uptime,
	}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%s/heartbeat", c.nodeID), body, &resp); err != nil {
		return false, 0, fmt.Errorf("heartbeat request: %w", err)
	}
	if resp.Code != 0 {
		return false, 0, fmt.Errorf("heartbeat failed: %s", resp.Message)
	}

	var data heartbeatResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return false, 0, fmt.Errorf("decode heartbeat response: %w", err)
	}
	return data.ConfigChanged, data.PullInterval, nil
}

// Report sends consolidated traffic, alive IP and status data to the panel.
func (c *Client) Report(traffic map[string][2]int64, alive map[string][]string, cpu, mem float64, uptime uint64) error {
	payload := map[string]interface{}{
		"traffic": traffic,
		"alive":   alive,
		"status": map[string]interface{}{
			"cpu":    cpu,
			"mem":    mem,
			"uptime": uptime,
		},
	}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%s/report", c.nodeID), payload, &resp); err != nil {
		return fmt.Errorf("report request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report failed: %s", resp.Message)
	}
	return nil
}

// ReportTraffic sends collected traffic statistics to the panel (Xboard-style format).
func (c *Client) ReportTraffic(data map[string][2]int64) error {
	if len(data) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"data": data,
	}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%s/traffic", c.nodeID), payload, &resp); err != nil {
		return fmt.Errorf("report traffic request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report traffic failed: %s", resp.Message)
	}
	return nil
}

// ReportAlive sends connected IP data to the panel.
func (c *Client) ReportAlive(data map[string][]string) error {
	body := map[string]interface{}{
		"data": data,
	}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%s/alive", c.nodeID), body, &resp); err != nil {
		return fmt.Errorf("report alive request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report alive failed: %s", resp.Message)
	}
	return nil
}

// FetchDeviceLimit retrieves per-user device limits from the panel.
func (c *Client) FetchDeviceLimit() (map[string]int, error) {
	var resp apiResponse
	if err := c.get("/api/internal/agent/devicelimit", &resp); err != nil {
		return nil, fmt.Errorf("fetch device limit request: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch device limit failed: %s", resp.Message)
	}

	var data deviceLimitResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode device limit response: %w", err)
	}
	if data.Limits == nil {
		data.Limits = make(map[string]int)
	}
	return data.Limits, nil
}

// post sends a JSON POST request to the panel.
func (c *Client) post(path string, payload interface{}, result *apiResponse) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Token", c.token)

	return c.doRequest(req, result)
}

// get sends a JSON GET request to the panel.
func (c *Client) get(path string, result *apiResponse) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Node-Token", c.token)

	return c.doRequest(req, result)
}

// doRequest executes an HTTP request and decodes the panel's standard response.
func (c *Client) doRequest(req *http.Request, result *apiResponse) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, string(body))
	}
	return nil
}
