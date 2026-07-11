package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"nexus-agent/internal/httpclient"
)

// connectionMetadata matches Clash / sing-box clash_api nested metadata.
type connectionMetadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationIP   string `json:"destinationIP"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	// Optional user identifiers (version-dependent)
	UID         string `json:"uid"`
	User        string `json:"user"`
	InboundUser string `json:"inboundUser"`
}

// connectionEntry represents a single active connection from Clash API.
type connectionEntry struct {
	ID            string              `json:"id"`
	Upload        int64               `json:"upload"`
	Download      int64               `json:"download"`
	Start         string              `json:"start"`
	Chains        []string            `json:"chains"`
	Rule          string              `json:"rule"`
	User          string              `json:"user"`
	InboundUser   string              `json:"inboundUser"`
	Source        string              `json:"source"`
	SourceIP      string              `json:"sourceIP"`
	Destination   string              `json:"destination"`
	DestinationIP string              `json:"destinationIP"`
	Network       string              `json:"network"`
	Type          string              `json:"type"`
	Inbound       string              `json:"inbound"`
	Metadata      *connectionMetadata `json:"metadata"`
}

// connectionsResponse is the Clash /connections response.
type connectionsResponse struct {
	Connections   []connectionEntry `json:"connections"`
	DownloadTotal int64             `json:"downloadTotal"`
	UploadTotal   int64             `json:"uploadTotal"`
}

// StatsCollector queries the sing-box Clash API and returns per-user traffic / alive data.
type StatsCollector struct {
	statsURL string
	client   *http.Client
	nodeID   string

	mu sync.Mutex
	// lastConnTraffic: connection id -> [upload, download] last seen cumulative
	lastConnTraffic map[string][2]int64
	// knownUsers: full UUIDs currently authorized on this node (for fallback / short-name map)
	knownUsers []string
	// shortToUUID: first-8-of-uuid / other name -> full UUID
	shortToUUID map[string]string
}

// New creates a new StatsCollector that queries the given Clash API base URL.
func New(statsURL string, nodeID string) *StatsCollector {
	return &StatsCollector{
		statsURL:        strings.TrimRight(statsURL, "/"),
		client:          &http.Client{Timeout: 10 * time.Second},
		nodeID:          nodeID,
		lastConnTraffic: make(map[string][2]int64),
		shortToUUID:     make(map[string]string),
	}
}

// SetKnownUsers updates the authorized user list for name→UUID resolution and single-user fallback.
func (s *StatsCollector) SetKnownUsers(uuids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.knownUsers = append([]string(nil), uuids...)
	s.shortToUUID = make(map[string]string, len(uuids)*2)
	for _, u := range uuids {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		s.shortToUUID[u] = u
		s.shortToUUID[strings.ToLower(u)] = u
		if len(u) >= 8 {
			s.shortToUUID[u[:8]] = u
			s.shortToUUID[strings.ToLower(u[:8])] = u
		}
		// hysteria2-style name: hex without dashes, first 8
		compact := strings.ReplaceAll(u, "-", "")
		if len(compact) >= 8 {
			s.shortToUUID[compact[:8]] = u
			s.shortToUUID[strings.ToLower(compact[:8])] = u
		}
		if compact != "" {
			s.shortToUUID[compact] = u
		}
	}
}

func (c connectionEntry) sourceIP() string {
	if c.Metadata != nil && c.Metadata.SourceIP != "" {
		return c.Metadata.SourceIP
	}
	if c.SourceIP != "" {
		return c.SourceIP
	}
	if host, _, err := net.SplitHostPort(c.Source); err == nil {
		return host
	}
	return c.Source
}

func (c connectionEntry) rawUser() string {
	if c.InboundUser != "" {
		return c.InboundUser
	}
	if c.User != "" {
		return c.User
	}
	if c.Metadata != nil {
		if c.Metadata.InboundUser != "" {
			return c.Metadata.InboundUser
		}
		if c.Metadata.User != "" {
			return c.Metadata.User
		}
		if c.Metadata.UID != "" {
			return c.Metadata.UID
		}
	}
	return ""
}

// isInboundProxyConn returns true for connections that entered via a proxy inbound
// (as opposed to pure local/control traffic with no useful accounting).
func (c connectionEntry) isInboundProxyConn() bool {
	typ := c.Type
	if c.Metadata != nil && c.Metadata.Type != "" {
		typ = c.Metadata.Type
	}
	typ = strings.ToLower(typ)
	// Clash metadata.type examples: "vless/vless-in", "hysteria2/hysteria2-in", "Socks5", ...
	for _, p := range []string{"vless", "hysteria", "tuic", "trojan", "shadowsocks", "vmess", "anytls"} {
		if strings.Contains(typ, p) {
			return true
		}
	}
	if c.Inbound != "" {
		in := strings.ToLower(c.Inbound)
		for _, p := range []string{"vless", "hysteria", "tuic", "trojan", "ss-", "vmess"} {
			if strings.Contains(in, p) {
				return true
			}
		}
	}
	return false
}

func (s *StatsCollector) resolveUser(raw string, isProxy bool) string {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if full, ok := s.shortToUUID[raw]; ok {
			return full
		}
		if full, ok := s.shortToUUID[strings.ToLower(raw)]; ok {
			return full
		}
		// Already a full UUID not in map (stale user still connected)
		if len(raw) >= 32 && strings.Count(raw, "-") >= 4 {
			return raw
		}
		return raw
	}
	// Clash often omits user on inbound connections; if only one known user, attribute to them.
	if isProxy && len(s.knownUsers) == 1 {
		return s.knownUsers[0]
	}
	return ""
}

func (s *StatsCollector) fetchConnections() (*connectionsResponse, error) {
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
	return &conns, nil
}

// Collect queries Clash /connections and returns delta traffic per user.
// Counters are tracked per connection id so closed connections still contribute
// their last observed delta before disappearing.
func (s *StatsCollector) Collect() ([]httpclient.TrafficEntry, error) {
	conns, err := s.fetchConnections()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Aggregate deltas by user UUID
	userDelta := make(map[string][2]int64)
	seenIDs := make(map[string]struct{}, len(conns.Connections))

	for _, c := range conns.Connections {
		if c.ID == "" {
			continue
		}
		seenIDs[c.ID] = struct{}{}
		isProxy := c.isInboundProxyConn()
		user := s.resolveUser(c.rawUser(), isProxy)
		if user == "" {
			continue
		}

		curUp, curDown := c.Upload, c.Download
		prev := s.lastConnTraffic[c.ID]
		var dUp, dDown int64
		if curUp >= prev[0] {
			dUp = curUp - prev[0]
		} else {
			dUp = curUp // counter reset
		}
		if curDown >= prev[1] {
			dDown = curDown - prev[1]
		} else {
			dDown = curDown
		}
		s.lastConnTraffic[c.ID] = [2]int64{curUp, curDown}

		if dUp == 0 && dDown == 0 {
			continue
		}
		agg := userDelta[user]
		userDelta[user] = [2]int64{agg[0] + dUp, agg[1] + dDown}
	}

	// Drop tracking for closed connections (their final bytes already counted)
	for id := range s.lastConnTraffic {
		if _, ok := seenIDs[id]; !ok {
			delete(s.lastConnTraffic, id)
		}
	}

	entries := make([]httpclient.TrafficEntry, 0, len(userDelta))
	for user, d := range userDelta {
		if d[0] == 0 && d[1] == 0 {
			continue
		}
		entries = append(entries, httpclient.TrafficEntry{
			NodeID:   s.nodeID,
			UserUUID: user,
			Upload:   d[0],
			Download: d[1],
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

// CollectAliveIPs queries /connections and extracts per-user source IPs.
func (s *StatsCollector) CollectAliveIPs() (map[string][]string, error) {
	conns, err := s.fetchConnections()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string][]string)
	seen := make(map[string]map[string]bool)

	for _, c := range conns.Connections {
		isProxy := c.isInboundProxyConn()
		user := s.resolveUser(c.rawUser(), isProxy)
		sourceIP := c.sourceIP()
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
