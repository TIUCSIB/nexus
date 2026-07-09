// Package ws implements a WebSocket hub for real-time communication
// between the panel and connected node/machine agents, following Nexus-Agent's
// WebSocket command pattern.
//
// Supported commands (panel → agent):
//   - restart      — Restart the proxy service (sing-box)
//   - reload       — Force reload config from panel
//   - update       — Update the kernel binary
//   - install      — Install the kernel binary
//   - sync.config  — Push full node config (agent applies via hot reload)
//   - sync.users   — Push full user list (agent applies via hot reload)
//
// Machine-level events:
//   - sync.nodes   — Notify machine that its node list has changed
//
// In machine mode, commands carry a "node_id" field in their data payload
// so the agent can dispatch to the correct node sub-process.
package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Non-browser clients (agents) don't send Origin
			return true
		}
		// Allow same-origin requests
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		selfOrigin := fmt.Sprintf("%s://%s", scheme, r.Host)
		if origin == selfOrigin {
			return true
		}
		// Also allow origin with port stripped
		selfOriginNoPort := fmt.Sprintf("%s://%s", scheme, strings.Split(r.Host, ":")[0])
		if origin == selfOriginNoPort {
			return true
		}
		log.Printf("[ws] rejected connection from origin: %s (expected: %s)", origin, selfOrigin)
		return false
	},
}

// Command represents a message sent from panel to agent.
type Command struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Message represents a message received from an agent.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// AgentConn represents a single WebSocket connection from an agent (node or machine).
type AgentConn struct {
	NodeID    string
	MachineID string // non-empty for machine-mode connections
	NodeIDs   []string // node IDs belonging to this machine connection
	conn      *websocket.Conn
	hub       *Hub
	send      chan []byte
	lastSeen  time.Time
	mu        sync.Mutex
}

// Hub manages all connected agent WebSocket connections.
type Hub struct {
	mu        sync.RWMutex
	agents    map[string]*AgentConn   // nodeID → agent (or nodeID → machine-connection)
	machines  map[string]*AgentConn   // machineID → machine connection
	commands  chan *CommandEnvelope
}

// CommandEnvelope wraps a command with the target node ID.
type CommandEnvelope struct {
	NodeID  string
	Command *Command
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		agents:   make(map[string]*AgentConn),
		machines: make(map[string]*AgentConn),
		commands: make(chan *CommandEnvelope, 100),
	}
}

// ServeWS handles an incoming WebSocket upgrade request.
// Supports two auth modes:
//   - node mode: node_id + server_token (legacy)
//   - machine mode: machine_id + machine_token (nexus-agent style)
//
// The agent must authenticate via token query param or X-Node-Token / X-Machine-Token header.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, nodeID, machineID, token string) {
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		token = r.Header.Get("X-Node-Token")
	}
	if token == "" {
		http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	// Token validation happens in the caller (middleware/handler knows which
	// token to validate). We store the token-validated identity here.

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	isMachine := machineID != ""

	agent := &AgentConn{
		NodeID:    nodeID,
		MachineID: machineID,
		conn:      conn,
		hub:       h,
		send:      make(chan []byte, 256),
		lastSeen:  time.Now(),
	}

	h.mu.Lock()
	if isMachine {
		// Machine mode: register under machine_id
		if existing, ok := h.machines[machineID]; ok {
			existing.close()
		}
		h.machines[machineID] = agent

		// Also register all node IDs belonging to this machine
		// (the handler sets NodeIDs before calling ServeWS)
		for _, nid := range agent.NodeIDs {
			if existing, ok := h.agents[nid]; ok {
				// Only close if it's the same machine connection type
				if existing.MachineID != machineID {
					existing.close()
				}
			}
			h.agents[nid] = agent
		}

		log.Printf("[ws] machine connected: machine=%s, nodes=%v", machineID, agent.NodeIDs)
	} else {
		// Node mode: register under node_id
		if existing, ok := h.agents[nodeID]; ok {
			existing.close()
		}
		h.agents[nodeID] = agent

		log.Printf("[ws] node connected: node=%s", nodeID)
	}
	h.mu.Unlock()

	// Send auth success
	authMsg, _ := json.Marshal(map[string]interface{}{
		"event": "auth.success",
		"data": map[string]interface{}{
			"node_id":    nodeID,
			"machine_id": machineID,
		},
	})
	select {
	case agent.send <- authMsg:
	default:
	}

	go agent.writePump()
	go agent.readPump()
}

// SendCommand sends a command to a specific agent, or to all agents if nodeID is "*".
// If the target node belongs to a machine connection, the command data is wrapped
// with a "node_id" field for agent-side dispatch.
func (h *Hub) SendCommand(nodeID string, cmd *Command) error {
	if nodeID == "*" {
		h.mu.RLock()
		defer h.mu.RUnlock()
		for _, agent := range h.agents {
			agent.sendCommand(cmd)
		}
		return nil
	}

	h.mu.RLock()
	agent, ok := h.agents[nodeID]
	h.mu.RUnlock()
	if !ok {
		return fmt.Errorf("agent %s not connected", nodeID)
	}

	// If this is a machine connection, wrap the command with node_id
	if agent.MachineID != "" {
		wrapped := &Command{
			Type: cmd.Type,
		}
		var dataMap map[string]interface{}
		if cmd.Data != nil {
			json.Unmarshal(cmd.Data, &dataMap)
		} else {
			dataMap = make(map[string]interface{})
		}
		dataMap["node_id"] = nodeID
		wrappedData, _ := json.Marshal(dataMap)
		wrapped.Data = wrappedData
		agent.sendCommand(wrapped)
		return nil
	}

	agent.sendCommand(cmd)
	return nil
}

// SendMachineCommand sends a machine-level event (e.g. sync.nodes) to a specific machine.
// If machineID is "*", sends to all connected machines.
func (h *Hub) SendMachineCommand(machineID string, cmd *Command) error {
	if machineID == "*" {
		h.mu.RLock()
		defer h.mu.RUnlock()
		for _, agent := range h.machines {
			agent.sendCommand(cmd)
		}
		return nil
	}

	h.mu.RLock()
	agent, ok := h.machines[machineID]
	h.mu.RUnlock()
	if !ok {
		return fmt.Errorf("machine %s not connected", machineID)
	}
	agent.sendCommand(cmd)
	return nil
}

// RegisterMachineNodes updates the node IDs for a machine connection.
// This is called after initial discovery or after a sync.nodes event.
func (h *Hub) RegisterMachineNodes(machineID string, nodeIDs []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	agent, ok := h.machines[machineID]
	if !ok {
		return
	}

	// Remove old node registrations
	oldNodes := agent.NodeIDs
	for _, nid := range oldNodes {
		// Only remove if still pointing to this machine's agent
		if existing, ok := h.agents[nid]; ok && existing.MachineID == machineID {
			delete(h.agents, nid)
		}
	}

	// Add new node registrations
	agent.NodeIDs = nodeIDs
	for _, nid := range nodeIDs {
		h.agents[nid] = agent
	}

	log.Printf("[ws] machine %s node list updated: %v", machineID, nodeIDs)
}

// NotifyMachineNodesChanged sends a sync.nodes event to trigger agent rediscovery.
func (h *Hub) NotifyMachineNodesChanged(machineID string) {
	data, _ := json.Marshal(map[string]interface{}{
		"event": "sync.nodes",
		"data": map[string]interface{}{
			"machine_id": machineID,
		},
	})
	cmd := &Command{
		Type: "sync.nodes",
		Data: data,
	}
	h.SendMachineCommand(machineID, cmd)
}

// PushConfig sends a sync.config event with the full node config to a connected agent.
// This allows the agent to apply config changes without HTTP polling.
func (h *Hub) PushConfig(nodeID string, config interface{}) {
	data, _ := json.Marshal(config)
	cmd := &Command{
		Type: "sync.config",
		Data: data,
	}
	if err := h.SendCommand(nodeID, cmd); err != nil {
		log.Printf("[ws] push config to %s failed: %v", nodeID, err)
	}
}

// PushUsers sends a sync.users event with the full user list to a connected agent.
// This allows the agent to apply user changes without HTTP polling.
func (h *Hub) PushUsers(nodeID string, users interface{}) {
	data, _ := json.Marshal(users)
	cmd := &Command{
		Type: "sync.users",
		Data: data,
	}
	if err := h.SendCommand(nodeID, cmd); err != nil {
		log.Printf("[ws] push users to %s failed: %v", nodeID, err)
	}
}

// Disconnect removes an agent from the hub and closes its connection.
func (h *Hub) Disconnect(nodeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent, ok := h.agents[nodeID]; ok {
		if agent.MachineID != "" {
			// Machine connection: remove all node registrations
			for _, nid := range agent.NodeIDs {
				delete(h.agents, nid)
			}
			delete(h.machines, agent.MachineID)
			agent.close()
			log.Printf("[ws] machine disconnected: machine=%s, nodes=%v", agent.MachineID, agent.NodeIDs)
		} else {
			delete(h.agents, nodeID)
			agent.close()
			log.Printf("[ws] node disconnected: node=%s", nodeID)
		}
	}
}

// IsConnected returns whether a node agent is currently connected.
func (h *Hub) IsConnected(nodeID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[nodeID]
	return ok
}

// IsMachineConnected returns whether a machine is currently connected.
func (h *Hub) IsMachineConnected(machineID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.machines[machineID]
	return ok
}

// Count returns the number of currently connected agents.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.agents)
}

// MachineCount returns the number of connected machines.
func (h *Hub) MachineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.machines)
}

func (a *AgentConn) sendCommand(cmd *Command) {
	data, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	select {
	case a.send <- data:
	default:
		log.Printf("[ws] send buffer full for %s, dropping command", a.NodeID)
	}
}

func (a *AgentConn) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		id := a.NodeID
		if a.MachineID != "" {
			id = "machine:" + a.MachineID
		}
		a.hub.Disconnect(id)
	}()

	for {
		select {
		case message, ok := <-a.send:
			if !ok {
				a.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			a.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := a.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[ws] write error for %s: %v", a.NodeID, err)
				a.hub.Disconnect(a.NodeID)
				return
			}
		case <-ticker.C:
			a.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := a.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (a *AgentConn) readPump() {
	defer a.hub.Disconnect(a.NodeID)
	a.conn.SetReadLimit(4096)
	a.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	a.conn.SetPongHandler(func(string) error {
		a.mu.Lock()
		a.lastSeen = time.Now()
		a.mu.Unlock()
		a.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := a.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error for %s: %v", a.NodeID, err)
			}
			return
		}

		a.mu.Lock()
		a.lastSeen = time.Now()
		a.mu.Unlock()

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "traffic":
			log.Printf("[ws] traffic data from %s: %s", a.NodeID, string(msg.Data))
		case "status":
			log.Printf("[ws] status from %s: %s", a.NodeID, string(msg.Data))
		case "pong":
		}
	}
}

func (a *AgentConn) close() {
	a.conn.Close()
}