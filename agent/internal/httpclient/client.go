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
	NodeID   uint   `json:"node_id"`
	UserUUID string `json:"user_uuid"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// heartbeatResponse is what the panel returns on heartbeat.
type heartbeatResponse struct {
	ConfigChanged bool `json:"config_changed"`
}

// singboxConfigResponse is what the panel returns on GET /internal/agent/:node_id/config.
type singboxConfigResponse struct {
	ConfigJSON string `json:"config_json"`
	UsersJSON  string `json:"users_json"`
	RoutesJSON string `json:"routes_json"`
}

// apiResponse is the generic wrapper returned by all panel endpoints.
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// deviceLimitResponse is what the panel returns on GET /internal/agent/devicelimit.
type deviceLimitResponse struct {
	Limits map[string]int `json:"limits"`
}

// aliveRequest is the payload sent to POST /internal/agent/:node_id/alive.
type aliveRequest struct {
	Data map[string][]string `json:"data"`
}

// Client communicates with the Nexus panel over HTTP REST.
type Client struct {
	baseURL    string
	token      string // global server_token
	nodeID     int
	httpClient *http.Client
}

// NewClient creates a new Client pointing at the given panel address.
func NewClient(baseURL, token string, nodeID int) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		nodeID:  nodeID,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Heartbeat sends periodic status to the panel and returns whether config changed.
func (c *Client) Heartbeat(cpu, mem float64, uptime uint64) (configChanged bool, err error) {
	body := map[string]interface{}{
		"cpu":    cpu,
		"mem":    mem,
		"uptime": uptime,
	}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%d/heartbeat", c.nodeID), body, &resp); err != nil {
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
func (c *Client) GetConfig() (configJSON string, usersJSON string, routesJSON string, err error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+fmt.Sprintf("/api/internal/agent/%d/config", c.nodeID), nil)
	if err != nil {
		return "", "", "", fmt.Errorf("create config request: %w", err)
	}
	req.Header.Set("X-Node-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("get config request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("read config response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", "", "", fmt.Errorf("decode config response: %w (body: %s)", err, string(body))
	}
	if apiResp.Code != 0 {
		return "", "", "", fmt.Errorf("get config failed: %s", apiResp.Message)
	}

	var data singboxConfigResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return "", "", "", fmt.Errorf("decode config data: %w", err)
	}

	return data.ConfigJSON, data.UsersJSON, data.RoutesJSON, nil
}

// ReportTraffic sends collected traffic statistics to the panel.
func (c *Client) ReportTraffic(entries []TrafficEntry) error {
	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%d/traffic", c.nodeID), entries, &resp); err != nil {
		return fmt.Errorf("report traffic request: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("report traffic failed: %s", resp.Message)
	}
	return nil
}

// ReportAlive sends connected IP data to the panel.
func (c *Client) ReportAlive(data map[string][]string) error {
	body := aliveRequest{Data: data}

	var resp apiResponse
	if err := c.post(fmt.Sprintf("/api/internal/agent/%d/alive", c.nodeID), body, &resp); err != nil {
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
