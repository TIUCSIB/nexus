package devicelimit

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type singboxConnection struct {
	ID            string `json:"id"`
	User          string `json:"user"`
	Upload        int64  `json:"upload"`
	Download      int64  `json:"download"`
	Start         string `json:"start"`
	Network       string `json:"network"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	SourceIP      string `json:"sourceIP"`
	Destination   string `json:"destination"`
	DestinationIP string `json:"destinationIP"`
	Inbound       string `json:"inbound"`
	InboundUser   string `json:"inboundUser"`
	Outbound      string `json:"outbound"`
}

type singboxConnectionsResponse struct {
	Connections []singboxConnection `json:"connections"`
}

func (c singboxConnection) userUUID() string {
	if c.InboundUser != "" {
		return c.InboundUser
	}
	return c.User
}

func (c singboxConnection) sourceIP() string {
	if c.SourceIP != "" {
		return c.SourceIP
	}
	if host, _, err := net.SplitHostPort(c.Source); err == nil {
		return host
	}
	return c.Source
}

type Enforcer struct {
	statsURL       string
	client         *http.Client
	deviceLimits   map[string]int
	mu             sync.RWMutex
	recentlyClosed map[string]time.Time
	closedEvents   uint64
}

func New(statsURL string) *Enforcer {
	return &Enforcer{
		statsURL:       statsURL,
		client:         &http.Client{Timeout: 10 * time.Second},
		deviceLimits:   make(map[string]int),
		recentlyClosed: make(map[string]time.Time),
	}
}

func (e *Enforcer) UpdateLimits(limits map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.deviceLimits = limits
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

	userMap := make(map[string]*userConns)
	for _, conn := range conns {
		uuid := conn.userUUID()
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
	e.mu.RLock()
	limits := e.deviceLimits
	e.mu.RUnlock()

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
