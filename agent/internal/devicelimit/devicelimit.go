package devicelimit

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type connectionMetadata struct {
	SourceIP    string `json:"sourceIP"`
	User        string `json:"user"`
	InboundUser string `json:"inboundUser"`
	UID         string `json:"uid"`
	Type        string `json:"type"`
}

type singboxConnection struct {
	ID            string              `json:"id"`
	User          string              `json:"user"`
	Upload        int64               `json:"upload"`
	Download      int64               `json:"download"`
	Start         string              `json:"start"`
	Network       string              `json:"network"`
	Type          string              `json:"type"`
	Source        string              `json:"source"`
	SourceIP      string              `json:"sourceIP"`
	Destination   string              `json:"destination"`
	DestinationIP string              `json:"destinationIP"`
	Inbound       string              `json:"inbound"`
	InboundUser   string              `json:"inboundUser"`
	Outbound      string              `json:"outbound"`
	Metadata      *connectionMetadata `json:"metadata"`
}

type singboxConnectionsResponse struct {
	Connections []singboxConnection `json:"connections"`
}

func (c singboxConnection) rawUser() string {
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

func (c singboxConnection) sourceIP() string {
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

func (c singboxConnection) isProxyInbound() bool {
	typ := c.Type
	if c.Metadata != nil && c.Metadata.Type != "" {
		typ = c.Metadata.Type
	}
	typ = strings.ToLower(typ)
	for _, p := range []string{"vless", "hysteria", "tuic", "trojan", "shadowsocks", "vmess", "anytls"} {
		if strings.Contains(typ, p) {
			return true
		}
	}
	return false
}

type Enforcer struct {
	statsURL       string
	client         *http.Client
	deviceLimits   map[string]int
	knownUsers     []string
	shortToUUID    map[string]string
	mu             sync.RWMutex
	recentlyClosed map[string]time.Time
	closedEvents   uint64
}

func New(statsURL string) *Enforcer {
	return &Enforcer{
		statsURL:       strings.TrimRight(statsURL, "/"),
		client:         &http.Client{Timeout: 10 * time.Second},
		deviceLimits:   make(map[string]int),
		shortToUUID:    make(map[string]string),
		recentlyClosed: make(map[string]time.Time),
	}
}

func (e *Enforcer) UpdateLimits(limits map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.deviceLimits = limits
}

// SetKnownUsers helps resolve short names / single-user fallback when Clash omits user field.
func (e *Enforcer) SetKnownUsers(uuids []string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.knownUsers = append([]string(nil), uuids...)
	e.shortToUUID = make(map[string]string, len(uuids)*2)
	for _, u := range uuids {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		e.shortToUUID[u] = u
		e.shortToUUID[strings.ToLower(u)] = u
		if len(u) >= 8 {
			e.shortToUUID[u[:8]] = u
			e.shortToUUID[strings.ToLower(u[:8])] = u
		}
		compact := strings.ReplaceAll(u, "-", "")
		if len(compact) >= 8 {
			e.shortToUUID[compact[:8]] = u
		}
	}
}

func (e *Enforcer) resolveUser(raw string, isProxy bool) string {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if full, ok := e.shortToUUID[raw]; ok {
			return full
		}
		if full, ok := e.shortToUUID[strings.ToLower(raw)]; ok {
			return full
		}
		return raw
	}
	if isProxy && len(e.knownUsers) == 1 {
		return e.knownUsers[0]
	}
	return ""
}

func (e *Enforcer) Enforce() (int, error) {
	e.cleanRecentlyClosed()
	conns, err := e.fetchConnections()
	if err != nil {
		return 0, fmt.Errorf("fetch connections: %w", err)
	}
	if len(conns) == 0 {
		return 0, nil
	}

	type userConns struct {
		conns []singboxConnection
		ips   map[string]bool
	}

	e.mu.RLock()
	limits := e.deviceLimits
	e.mu.RUnlock()

	userMap := make(map[string]*userConns)
	for _, conn := range conns {
		e.mu.RLock()
		uuid := e.resolveUser(conn.rawUser(), conn.isProxyInbound())
		e.mu.RUnlock()
		sourceIP := conn.sourceIP()
		if uuid == "" || sourceIP == "" {
			continue
		}
		if userMap[uuid] == nil {
			userMap[uuid] = &userConns{ips: make(map[string]bool)}
		}
		userMap[uuid].conns = append(userMap[uuid].conns, conn)
		userMap[uuid].ips[sourceIP] = true
	}

	var closed int
	for uuid, uc := range userMap {
		limit, ok := limits[uuid]
		if !ok || limit <= 0 {
			continue
		}
		ipCount := len(uc.ips)
		if ipCount <= limit {
			continue
		}
		excess := ipCount - limit
		toClose := e.selectConnectionsToClose(uc.conns, excess)
		for _, conn := range toClose {
			e.mu.Lock()
			if _, recently := e.recentlyClosed[conn.ID]; recently {
				e.mu.Unlock()
				continue
			}
			e.recentlyClosed[conn.ID] = time.Now()
			e.mu.Unlock()
			if err := e.closeConnection(conn.ID); err != nil {
				log.Printf("[devicelimit] close conn %s for user %s failed: %v", conn.ID[:8], uuid[:8], err)
				continue
			}
			closed++
			log.Printf("[devicelimit] closed conn %s (user=%s, src=%s)", conn.ID[:8], uuid[:8], conn.sourceIP())
		}
	}
	if closed > 0 {
		e.mu.Lock()
		e.closedEvents += uint64(closed)
		e.mu.Unlock()
	}
	return closed, nil
}

func (e *Enforcer) selectConnectionsToClose(conns []singboxConnection, excess int) []singboxConnection {
	seen := make(map[string]bool)
	var result []singboxConnection
	for _, c := range conns {
		if excess <= 0 {
			break
		}
		sourceIP := c.sourceIP()
		if sourceIP == "" {
			continue
		}
		if !seen[sourceIP] {
			seen[sourceIP] = true
			result = append(result, c)
			excess--
		}
	}
	return result
}

func (e *Enforcer) fetchConnections() ([]singboxConnection, error) {
	url := e.statsURL + "/connections"
	resp, err := e.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("query sing-box connections: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sing-box connections HTTP %d: %s", resp.StatusCode, string(body))
	}
	var data singboxConnectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode sing-box connections: %w", err)
	}
	return data.Connections, nil
}

func (e *Enforcer) closeConnection(connID string) error {
	url := e.statsURL + "/connections/" + connID
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (e *Enforcer) cleanRecentlyClosed() {
	e.mu.Lock()
	defer e.mu.Unlock()
	cutoff := time.Now().Add(-60 * time.Second)
	for id, t := range e.recentlyClosed {
		if t.Before(cutoff) {
			delete(e.recentlyClosed, id)
		}
	}
}

func (e *Enforcer) ClosedEvents() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.closedEvents
}

func (e *Enforcer) HasLimits() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.deviceLimits) > 0
}
