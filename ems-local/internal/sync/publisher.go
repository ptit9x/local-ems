package sync

import (
	"sync"
	"time"
)

// Publisher sends messages to the cloud.
type Publisher interface {
	Connect() error
	Publish(topic string, payload []byte) error
	IsConnected() bool
	Close() error
}

// StubPublisher is a no-op publisher for testing and when sync is disabled.
type StubPublisher struct {
	connected bool
	messages  []PublishedMessage
	mu        sync.Mutex
}

// PublishedMessage represents a message published by the stub.
type PublishedMessage struct {
	Topic   string
	Payload []byte
	Time    time.Time
}

// NewStubPublisher creates a new StubPublisher.
func NewStubPublisher() *StubPublisher {
	return &StubPublisher{}
}

// Connect implements Publisher.
func (p *StubPublisher) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = true
	return nil
}

// Publish implements Publisher.
func (p *StubPublisher) Publish(topic string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil
	}
	p.messages = append(p.messages, PublishedMessage{
		Topic:   topic,
		Payload: append([]byte(nil), payload...), // copy
		Time:    time.Now().UTC(),
	})
	return nil
}

// IsConnected implements Publisher.
func (p *StubPublisher) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connected
}

// Close implements Publisher.
func (p *StubPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = false
	return nil
}

// GetMessages returns the published messages.
func (p *StubPublisher) GetMessages() []PublishedMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.messages
}
