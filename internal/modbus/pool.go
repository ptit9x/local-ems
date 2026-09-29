// Package modbus provides a Modbus TCP connection pool for multi-device communication.
package modbus

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnPool manages persistent Modbus TCP connections to multiple devices.
type ConnPool struct {
	clients map[string]*poolEntry // key = "host:port/unitID"
	mu      sync.RWMutex
	timeout time.Duration
}

type poolEntry struct {
	client  *Client
	key     string
	host    string
	port    int
	unitID  uint8
	mu      sync.Mutex // per-connection lock for serial access
	healthy bool
	lastErr error
}

// NewConnPool creates a new connection pool.
func NewConnPool(timeout time.Duration) *ConnPool {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &ConnPool{
		clients: make(map[string]*poolEntry),
		timeout: timeout,
	}
}

// poolKey returns a unique key for a device.
func poolKey(host string, port int, unitID uint8) string {
	return fmt.Sprintf("%s:%d/%d", host, port, unitID)
}

// Register adds a device to the pool. Does not connect yet.
func (p *ConnPool) Register(host string, port int, unitID uint8) {
	key := poolKey(host, port, unitID)
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.clients[key]; exists {
		return
	}

	p.clients[key] = &poolEntry{
		key:    key,
		host:   host,
		port:   port,
		unitID: unitID,
	}
}

// Connect establishes TCP connections to all registered devices.
func (p *ConnPool) Connect() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var firstErr error
	for _, entry := range p.clients {
		if err := p.connectEntry(entry); err != nil {
			log.Printf("modbus pool: failed to connect %s: %v", entry.key, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (p *ConnPool) connectEntry(entry *poolEntry) error {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.client != nil {
		entry.client.Close()
	}

	addr := fmt.Sprintf("%s:%d", entry.host, entry.port)
	client := NewClient(addr, p.timeout)
	if err := client.Connect(); err != nil {
		entry.healthy = false
		entry.lastErr = err
		return err
	}

	entry.client = client
	entry.healthy = true
	entry.lastErr = nil
	return nil
}

// Get returns the client for a device. Thread-safe: the returned client
// should be used with the Do() method which holds the per-connection lock.
func (p *ConnPool) Get(host string, port int, unitID uint8) (*Client, error) {
	key := poolKey(host, port, unitID)
	p.mu.RLock()
	entry, ok := p.clients[key]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("modbus pool: device %s not registered", key)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.client == nil || !entry.healthy {
		// Try reconnect
		addr := fmt.Sprintf("%s:%d", entry.host, entry.port)
		client := NewClient(addr, p.timeout)
		if err := client.Connect(); err != nil {
			entry.healthy = false
			entry.lastErr = err
			return nil, fmt.Errorf("modbus pool: reconnect %s: %w", key, err)
		}
		entry.client = client
		entry.healthy = true
		entry.lastErr = nil
	}

	return entry.client, nil
}

// Do executes a function with the client for a device, holding the per-connection lock.
// This ensures serial Modbus access per TCP connection.
func (p *ConnPool) Do(host string, port int, unitID uint8, fn func(c *Client) error) error {
	key := poolKey(host, port, unitID)
	p.mu.RLock()
	entry, ok := p.clients[key]
	p.mu.RUnlock()

	if !ok {
		return fmt.Errorf("modbus pool: device %s not registered", key)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.client == nil || !entry.healthy {
		addr := fmt.Sprintf("%s:%d", entry.host, entry.port)
		client := NewClient(addr, p.timeout)
		if err := client.Connect(); err != nil {
			entry.healthy = false
			entry.lastErr = err
			return fmt.Errorf("modbus pool: reconnect %s: %w", key, err)
		}
		entry.client = client
		entry.healthy = true
	}

	err := fn(entry.client)
	if err != nil {
		entry.lastErr = err
		// Mark unhealthy on connection errors to trigger reconnect next time
		entry.healthy = false
	}
	return err
}

// HealthCheck returns the health status of all connections.
func (p *ConnPool) HealthCheck() map[string]HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]HealthStatus, len(p.clients))
	for key, entry := range p.clients {
		entry.mu.Lock()
		result[key] = HealthStatus{
			Connected: entry.healthy,
			LastError: entry.lastErr,
		}
		entry.mu.Unlock()
	}
	return result
}

// HealthStatus represents the health of a single connection.
type HealthStatus struct {
	Connected bool
	LastError error
}

// Close closes all connections in the pool.
func (p *ConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, entry := range p.clients {
		entry.mu.Lock()
		if entry.client != nil {
			entry.client.Close()
			entry.client = nil
		}
		entry.healthy = false
		entry.mu.Unlock()
	}
}

// Count returns the number of registered devices.
func (p *ConnPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

// HealthyCount returns the number of healthy connections.
func (p *ConnPool) HealthyCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, entry := range p.clients {
		entry.mu.Lock()
		if entry.healthy {
			count++
		}
		entry.mu.Unlock()
	}
	return count
}
