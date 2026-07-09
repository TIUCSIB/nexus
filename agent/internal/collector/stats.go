package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"nexus-agent/internal/httpclient"
)

// userStatsEntry matches the sing-box stats API response format.
type userStatsEntry struct {
	User     string `json:"user"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// statsResponse is the sing-box stats API response shape.
type statsResponse struct {
	Users []userStatsEntry `json:"users"`
}

// connectionEntry represents a single active connection from sing-box.
type connectionEntry struct {
	User        string `json:"user"`
	InboundUser string `json:"inboundUser"`
	Source      string `json:"source"`
	SourceIP    string `json:"sourceIP"`
	Destination string `json:"destination"`
	Network     string `json:"network"`
}

// connectionsResponse is the sing-box connections API response shape.
type connectionsResponse struct {
	Connections []connectionEntry `json:"connections"`
}

func connectionUser(c connectionEntry) string {
	if c.InboundUser != "" {
		return c.InboundUser
	}
	return c.User
}

func connectionSourceIP(c connectionEntry) string {
	if c.SourceIP != "" {
		return c.SourceIP
	}
	if host, _, err := net.SplitHostPort(c.Source); err == nil {
		return host
	}
	return c.Source
}

// StatsCollector queries the sing-box statistics API and returns per-user traffic data.
type StatsCollector struct {
	statsURL   string
	client     *http.Client
	nodeID     string
	lastTraffic map[string][2]uint64 // user -> [upload, download] cumulative
}

// New creates a new StatsCollector that queries the given sing-box stats URL.
func New(statsURL string, nodeID string) *StatsCollector {
	return &StatsCollector{
		statsURL: statsURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		nodeID:      nodeID,
		lastTraffic: make(map[string][2]uint64),
	}
}

// Collect queries the sing-box stats API and returns delta traffic entries for all users.
// It computes the difference from the last collected values so only incremental traffic is reported.
func (s *StatsCollector) Collect() ([]httpclient.TrafficEntry, error) {
	url := s.statsURL + "/api/v1/stats/user"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("query sing-box stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sing-box stats returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var stats statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode sing-box stats: %w", err)
	}

	entries := make([]httpclient.TrafficEntry, 0, len(stats.Users))
	for _, u := range stats.Users {
		currentUpload := uint64(u.Upload)
		currentDownload := uint64(u.Download)

		prev := s.lastTraffic[u.User]
		prevUpload := prev[0]
		prevDownload := prev[1]

		// Calculate delta; handle counter reset gracefully
		var deltaUpload, deltaDownload uint64
		if currentUpload >= prevUpload {
			deltaUpload = currentUpload - prevUpload
		} else {
			deltaUpload = currentUpload // counter reset, report current value
		}
		if currentDownload >= prevDownload {
			deltaDownload = currentDownload - prevDownload
		} else {
			deltaDownload = currentDownload
		}

		// Update stored cumulative values
		s.lastTraffic[u.User] = [2]uint64{currentUpload, currentDownload}

		// Skip users with no traffic delta
		if deltaUpload == 0 && deltaDownload == 0 {
			continue
		}

		entries = append(entries, httpclient.TrafficEntry{
			NodeID:   s.nodeID,
			UserUUID: u.User,
			Upload:   int64(deltaUpload),
			Download: int64(deltaDownload),
		})
	}

	return entries, nil
}

// CollectXboard returns traffic data in Xboard format: {"user_uuid": [upload, download]}
func (s *StatsCollector) CollectXboard() (map[string][2]int64, error) {
	entries, err := s.Collect()
	if err != nil {
		return nil, err
	}

	result := make(map[string][2]int64, len(entries))
	for _, e := range entries {
		result[e.UserUUID] = [2]int64{e.Upload, e.Download}
	}
	return result, nil
}

// CollectAliveIPs queries the sing-box connections API and extracts per-user source IPs.
// Returns a map of user UUID to list of unique source IPs.
func (s *StatsCollector) CollectAliveIPs() (map[string][]string, error) {
	url := s.statsURL + "/connections"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("query sing-box connections: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sing-box connections returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var conns connectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&conns); err != nil {
		return nil, fmt.Errorf("decode sing-box connections: %w", err)
	}

	result := make(map[string][]string)
	seen := make(map[string]map[string]bool) // user -> set of IPs

	for _, c := range conns.Connections {
		user := connectionUser(c)
		sourceIP := connectionSourceIP(c)
		if user == "" || sourceIP == "" {
			continue
		}
		if seen[user] == nil {
			seen[user] = make(map[string]bool)
		}
		if !seen[user][sourceIP] {
			seen[user][sourceIP] = true
			result[user] = append(result[user], sourceIP)
		}
	}

	return result, nil
}
