package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nexus-agent/internal/config"
)

// TrafficEntry represents a single user traffic record to report to the panel.
type TrafficEntry struct {
	NodeID   uint   `json:"node_id"`
	UserUUID string `json:"user_uuid"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// registerRequest is the payload sent to POST /internal/agent/register.
type registerRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Token   string `json:"token"`
}

// registerResponse is what the panel returns on successful registration.
type registerResponse struct {
	NodeID    uint   `json:"node_id"`
	AuthToken string `json:"auth_token"`
}

// heartbeatRequest is the payload sent to POST /internal/agent/heartbeat.
type heartbeatRequest struct {
	CPU    float64 `json:"cpu"`
	Mem    float64 `json:"mem"`
	Uptime uint64  `json:"uptime"`
}

// heartbeatResponse is what the panel returns on heartbeat.
type heartbeatResponse struct {
	ConfigChanged bool `json:"config_changed"`
}

// singboxConfigResponse is what the panel returns on GET /internal/agent/config.
type singboxConfigResponse struct {
	ConfigJSON string `json:"config_json"`
	UsersJSON  string `json:"users_json"`
}

// aliveRequest is the payload sent to POST /internal/agent/alive.
type aliveRequest struct {
	NodeID uint     `json:"node_id"`
	Data   map[string][]string `json:"data"`
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
	httpClient *http.Client
	configETag string // stored ETag for config caching
}

// NewClient creates a new Client pointing at the given panel address.
func NewClient(cfg config.PanelConfig) *Client {
	return &Client{
		baseURL: cfg.Address,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Register tells the panel about this node and receives an auth token.
func (c *Client) Register(name, address, registerToken string) (nodeID uint, authToken string, err error) {
	body := registerRequest{
		Name:    name,
		Address: address,
		Token:   registerToken,
	}

	var resp apiResponse
	if err := c.post("/api/v1/internal/agent/register", "", body, &resp); err != nil {
		return 0, "", fmt.Errorf("register request: %w", err)
	}
	if resp.Code != 0 {
		return 0, "", fmt.Errorf("register failed: %s", resp.Message)
	}

	var data registerResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return 0, "", fmt.Errorf("decode register response: %w", err)
	}
	return data.NodeID, data.AuthToken, nil
}

// Heartbeat sends periodic status to the panel and returns whether config changed.
func (c *Client) Heartbeat(authToken string, cpu, mem float64, uptime uint64) (configChanged bool, err error) {
	body := heartbeatRequest{
		CPU:    cpu,
		Mem:    mem,
		Uptime: uptime,
	}

	var resp apiResponse
	if err := c.post("/api/v1/internal/agent/heartbeat", authToken, body, &resp); err != nil {
		return false, fmt.Errorf("heartbeat request: %w", err)
	}
	if resp.Code != 0 {
		return false, fmt.Errorf("heartbeat failed: %s", resp.Message)
	}

	var data heartbeatResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return false, fmt.Errorf("decode heartbeat response: %w", err)
	}
	return data.ConfigChanged, nil
}

// GetConfig fetches the sing-box configuration for this node from the panel.
// Uses ETag caching: sends If-None-Match header, and on 304 returns empty strings
// with changed=false. On 200, updates the stored ETag and returns changed=true.
func (c *Client) GetConfig(authToken string) (configJSON string, usersJSON string, changed bool, err error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/internal/agent/config", nil)
	if err != nil {
		return "", "", false, fmt.Errorf("create config request: %w", err)
	}
	if authToken != "" {
		req.Header.Set("X-Node-Token", authToken)
	}
	if c.configETag != "" {
		req.Header.Set("If-None-Match", c.configETag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", false, fmt.Errorf("get config request: %w", err)
	}
	defer resp.Body.Close()

	// 304 Not Modified: config has not changed
	if resp.StatusCode == http.StatusNotModified {
		return "", "", false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", false, fmt.Errorf("read config response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", "", false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", "", false, fmt.Errorf("decode config response: %w (body: %s)", err, string(body))
	}
	if apiResp.Code != 0 {
		return "", "", false, fmt.Errorf("get config failed: %s", apiResp.Message)
	}

	var data singboxConfigResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return "", "", false, fmt.Errorf("decode config data: %w", err)
	}

	// Update stored ETag from response
	if etag := resp.Header.Get("ETag"); etag != "" {
		c.configETag = etag
	}

	return data.ConfigJSON, data.UsersJSON, true, nil
}

// ReportTraffic sends collected traffic statistics to the panel.
func (c *Client) ReportTraffic(authToken string, entries []TrafficEntry) error {
	var resp apiResponse
	if err := c.post("/api/v1/internal/agent/traffic", authToken, entries, &resp); err != nil {
		return fmt.Errorf("report traffic request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report traffic failed: %s", resp.Message)
	}
	return nil
}

// ReportAlive sends connected IP data to the panel.
func (c *Client) ReportAlive(nodeID uint, authToken string, data map[string][]string) error {
	body := aliveRequest{
		NodeID: nodeID,
		Data:   data,
	}

	var resp apiResponse
	if err := c.post("/api/v1/internal/agent/alive", authToken, body, &resp); err != nil {
		return fmt.Errorf("report alive request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report alive failed: %s", resp.Message)
	}
	return nil
}

// post sends a JSON POST request to the panel.
func (c *Client) post(path, authToken string, payload interface{}, result *apiResponse) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("X-Node-Token", authToken)
	}

	return c.doRequest(req, result)
}

// get sends a JSON GET request to the panel.
func (c *Client) get(path, authToken string, result *apiResponse) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if authToken != "" {
		req.Header.Set("X-Node-Token", authToken)
	}

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
