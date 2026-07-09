// Package wsclient implements the WebSocket client that connects from the
// node agent to the panel for real-time command reception.
//
// Reference: Xboard-node WebSocket command system
//   - After handshake, the agent connects to the panel's WS endpoint
//   - The panel can push commands: restart, reload, update, install
//   - The agent executes the command and reports back
package wsclient

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Command represents a command received from the panel.
type Command struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// CommandHandler is a function that handles a specific command type.
type CommandHandler func(cmd Command) error

// Client manages the WebSocket connection to the panel.
type Client struct {
	wsURL        string
	token        string
	nodeID       string
	conn         *websocket.Conn
	done         chan struct{}
	handlers     map[string]CommandHandler
	reconnectCh  chan struct{}
	connected    atomic.Bool
	OnReconnect  func() // called after successful reconnection
}

// IsConnected returns true if the WebSocket connection is active.
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// NewClient creates a new WebSocket client.
func NewClient(wsURL, token, nodeID string) *Client {
	return &Client{
		wsURL:       wsURL,
		token:       token,
		nodeID:      nodeID,
		done:        make(chan struct{}),
		handlers:    make(map[string]CommandHandler),
		reconnectCh: make(chan struct{}, 1),
	}
}

// NewMachineClient creates a new WebSocket client for machine mode.
// Uses machine_id + token for WS auth instead of node_id + token.
func NewMachineClient(wsURL, token string, machineID int) *Client {
	return &Client{
		wsURL:       wsURL,
		token:       token,
		nodeID:      fmt.Sprintf("machine:%d", machineID),
		done:        make(chan struct{}),
		handlers:    make(map[string]CommandHandler),
		reconnectCh: make(chan struct{}, 1),
	}
}

// RegisterHandler registers a handler for a command type.
func (c *Client) RegisterHandler(cmdType string, handler CommandHandler) {
	c.handlers[cmdType] = handler
}

// Connect establishes the WebSocket connection to the panel (node mode).
func (c *Client) Connect() error {
	u, err := url.Parse(c.wsURL)
	if err != nil {
		return fmt.Errorf("parse ws url: %w", err)
	}
	q := u.Query()
	q.Set("token", c.token)
	q.Set("node_id", c.nodeID)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial ws: %w", err)
	}
	c.conn = conn
	c.connected.Store(true)
	log.Printf("[ws] connected to panel at %s", u.Host+u.Path)

	// Start read pump
	go c.readPump()
	// Start ping sender
	go c.pingSender()

	return nil
}

// ConnectMachine establishes the WebSocket connection for machine mode.
// Uses machine_id query param instead of node_id for auth.
func (c *Client) ConnectMachine() error {
	u, err := url.Parse(c.wsURL)
	if err != nil {
		return fmt.Errorf("parse ws url: %w", err)
	}
	q := u.Query()
	q.Set("token", c.token)
	var machineID string
	fmt.Sscanf(c.nodeID, "machine:%s", &machineID)
	q.Set("machine_id", machineID)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial ws: %w", err)
	}
	c.conn = conn
	c.connected.Store(true)
	log.Printf("[ws] machine connected to panel at %s", u.Host+u.Path)

	go c.readPump()
	go c.pingSender()

	return nil
}

// Disconnect closes the WebSocket connection.
func (c *Client) Disconnect() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

// readPump reads incoming messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.connected.Store(false)
		c.conn.Close()
		// Attempt reconnection
		select {
		case c.reconnectCh <- struct{}{}:
		default:
		}
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error: %v", err)
			}
			return
		}

		var cmd Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("[ws] failed to parse command: %v", err)
			continue
		}

		log.Printf("[ws] received command: %s", cmd.Type)

		// Dispatch to registered handler
		if handler, ok := c.handlers[cmd.Type]; ok {
			if err := handler(cmd); err != nil {
				log.Printf("[ws] command %s failed: %v", cmd.Type, err)
			}
		} else {
			log.Printf("[ws] no handler for command type: %s", cmd.Type)
		}
	}
}

// pingSender sends periodic pings to keep the connection alive.
func (c *Client) pingSender() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[ws] ping error: %v", err)
				return
			}
		}
	}
}

// ReconnectLoop attempts to reconnect when the connection drops.
// Uses exponential backoff with jitter: 5s → 10s → 20s → 40s → 60s (max).
// Resets backoff on successful connection.
func (c *Client) ReconnectLoop() {
	const (
		initialBackoff = 5 * time.Second
		maxBackoff     = 60 * time.Second
		backoffFactor  = 2
	)
	backoff := initialBackoff

	for range c.reconnectCh {
		// Add ±20% random jitter to prevent thundering herd
		jitter := time.Duration(rand.Int63n(int64(backoff)/5) - int64(backoff)/10)
		delay := backoff + jitter
		if delay < time.Second {
			delay = time.Second
		}
		log.Printf("[ws] reconnecting in %v...", delay)
		time.Sleep(delay)

		// Re-create done channel
		c.done = make(chan struct{})

		if err := c.Connect(); err != nil {
			log.Printf("[ws] reconnection failed: %v", err)
			backoff = backoff * backoffFactor
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			// Try again
			select {
			case c.reconnectCh <- struct{}{}:
			default:
			}
		} else {
			log.Printf("[ws] reconnected successfully")
			backoff = initialBackoff // reset on success
			if c.OnReconnect != nil {
				c.OnReconnect()
			}
		}
	}
}