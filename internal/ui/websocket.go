package ui

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"sync"

	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
)

// maxWSClients limits concurrent WebSocket connections to prevent resource exhaustion.
const maxWSClients = 100

// WSHub manages WebSocket connections and broadcasts messages.
type WSHub struct {
	clients map[*wsConn]bool
	mu      sync.RWMutex
	done    chan struct{}
}

// wsConn is a minimal WebSocket connection (RFC 6455).
type wsConn struct {
	conn net.Conn
	bw   *bufio.Writer
	mu   sync.Mutex
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*wsConn]bool),
		done:    make(chan struct{}),
	}
}

// Run starts the hub (no-op for now, broadcasts are push-based).
func (h *WSHub) Run() {
	<-h.done
}

// Stop closes the hub.
func (h *WSHub) Stop() {
	close(h.done)
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.conn.Close()
	}
}

// Broadcast sends a message to all connected clients.
func (h *WSHub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		if err := c.writeMessage(msg); err != nil {
			go h.removeClient(c)
		}
	}
}

func (h *WSHub) addClient(c *wsConn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.clients) >= maxWSClients {
		return false
	}
	h.clients[c] = true
	return true
}

func (h *WSHub) removeClient(c *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		c.conn.Close()
		delete(h.clients, c)
	}
}

// HandleWS handles the WebSocket upgrade handshake and registers the connection.
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	// WebSocket handshake (RFC 6455)
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "not a websocket request", http.StatusBadRequest)
		return
	}

	acceptKey := computeAcceptKey(key)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket not supported", http.StatusInternalServerError)
		return
	}

	conn, brw, err := hj.Hijack()
	if err != nil {
		log.Printf("websocket hijack: %v", err)
		return
	}

	bw := brw.Writer

	// Send upgrade response
	resp := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)

	if _, err := bw.WriteString(resp); err != nil {
		conn.Close()
		return
	}
	bw.Flush()

	ws := &wsConn{conn: conn, bw: bw}
	if !h.addClient(ws) {
		conn.Close()
		return
	}

	// Read loop (to detect disconnects)
	go func() {
		defer h.removeClient(ws)
		buf := make([]byte, 512)
		for {
			if _, err := conn.Read(buf); err != nil {
				return
			}
		}
	}()
}

// writeMessage sends a WebSocket text frame (opcode 0x81).
func (c *wsConn) writeMessage(payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	length := len(payload)

	// Frame: FIN + Text opcode
	c.bw.WriteByte(0x81) // FIN=1, opcode=1 (text)

	// Payload length
	if length < 126 {
		c.bw.WriteByte(byte(length))
	} else if length < 65536 {
		c.bw.WriteByte(126)
		lenBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBytes, uint16(length))
		c.bw.Write(lenBytes)
	} else {
		c.bw.WriteByte(127)
		lenBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(lenBytes, uint64(length))
		c.bw.Write(lenBytes)
	}

	c.bw.Write(payload)
	return c.bw.Flush()
}

// computeAcceptKey generates the Sec-WebSocket-Accept header value.
func computeAcceptKey(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	io.WriteString(h, key+magic)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
