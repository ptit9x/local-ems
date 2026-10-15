package failover

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
)

type FailoverState int

const (
	FailoverNormal FailoverState = iota // Local EMS in control
	FailoverActive                      // Cloud has taken over
	FailoverRecovery                    // Transitioning back to Local
)

func (s FailoverState) String() string {
	switch s {
	case FailoverNormal:
		return "normal"
	case FailoverActive:
		return "active"
	case FailoverRecovery:
		return "recovery"
	default:
		return "unknown"
	}
}

type FailoverManager struct {
	mu          sync.RWMutex
	state       FailoverState
	siteID      string
	publisher   emsSync.Publisher
	topicPrefix string
	onFailover  func(state FailoverState)
}

func NewFailoverManager(pub emsSync.Publisher, siteID, topicPrefix string) *FailoverManager {
	return &FailoverManager{
		state:       FailoverNormal,
		siteID:      siteID,
		publisher:   pub,
		topicPrefix: topicPrefix,
	}
}

func (m *FailoverManager) State() FailoverState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *FailoverManager) setState(s FailoverState) {
	m.mu.Lock()
	if m.state == s {
		m.mu.Unlock()
		return
	}
	m.state = s
	fn := m.onFailover
	m.mu.Unlock()
	
	if fn != nil {
		fn(s)
	}
}

func (m *FailoverManager) OnStateChange(fn func(FailoverState)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFailover = fn
}

type recoveryRequest struct {
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *FailoverManager) RequestRecovery() error {
	m.setState(FailoverRecovery)

	req := recoveryRequest{
		Action:    "request_recovery",
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal recovery request: %w", err)
	}

	topic := m.topicPrefix + "/" + m.siteID + "/failover/recovery"
	if err := m.publisher.Publish(topic, data); err != nil {
		return fmt.Errorf("failed to publish recovery request: %w", err)
	}
	return nil
}

func (m *FailoverManager) AcknowledgeRecovery() {
	m.setState(FailoverNormal)
}

func (m *FailoverManager) HandleCloudMessage(action string, payload []byte) {
	switch action {
	case "takeover":
		m.setState(FailoverActive)
	case "handover_complete":
		m.AcknowledgeRecovery()
	}
}
