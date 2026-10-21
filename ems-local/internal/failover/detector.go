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
	LocalActive FailoverState = iota
	CloudTakeoverPending
	CloudActive
	HandoverPending
	FailoverFault
	// Backward-compatible names retained for callers while the v2 contract rolls out.
	FailoverNormal   = LocalActive
	FailoverActive   = CloudActive
	FailoverRecovery = HandoverPending
)

func (s FailoverState) String() string {
	switch s {
	case LocalActive:
		return "LOCAL_ACTIVE"
	case CloudTakeoverPending:
		return "CLOUD_TAKEOVER_PENDING"
	case CloudActive:
		return "CLOUD_ACTIVE"
	case HandoverPending:
		return "HANDOVER_PENDING"
	case FailoverFault:
		return "FAULT"
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
	leaseEpoch  uint64
}

func NewFailoverManager(pub emsSync.Publisher, siteID, topicPrefix string) *FailoverManager {
	return &FailoverManager{
		state:       LocalActive,
		siteID:      siteID,
		publisher:   pub,
		topicPrefix: topicPrefix,
	}
}

func (m *FailoverManager) LeaseEpoch() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.leaseEpoch
}

func (m *FailoverManager) ControlOwner() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.state == CloudActive || m.state == CloudTakeoverPending {
		return "cloud"
	}
	if m.state == FailoverFault || m.state == HandoverPending {
		return "none"
	}
	return "local"
}

// BeginStartupRecovery prevents a restarted Local EMS from issuing commands
// until Cloud confirms the current lease epoch or hands control back.
func (m *FailoverManager) BeginStartupRecovery() { m.setState(HandoverPending) }

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
	SchemaVersion int       `json:"schema_version"`
	SiteID        string    `json:"site_id"`
	Action        string    `json:"action"`
	Timestamp     time.Time `json:"timestamp"`
	LeaseEpoch    uint64    `json:"lease_epoch"`
}

func (m *FailoverManager) RequestRecovery() error {
	m.setState(FailoverRecovery)

	req := recoveryRequest{
		SchemaVersion: 2,
		SiteID:        m.siteID,
		Action:        "request_recovery",
		Timestamp:     time.Now().UTC(),
		LeaseEpoch:    m.LeaseEpoch(),
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
	var envelope struct {
		LeaseEpoch uint64 `json:"lease_epoch"`
	}
	_ = json.Unmarshal(payload, &envelope)
	m.mu.Lock()
	if envelope.LeaseEpoch < m.leaseEpoch {
		m.mu.Unlock()
		return
	}
	if envelope.LeaseEpoch > m.leaseEpoch {
		m.leaseEpoch = envelope.LeaseEpoch
	}
	m.mu.Unlock()
	switch action {
	case "takeover":
		m.setState(CloudActive)
	case "takeover_pending":
		m.setState(CloudTakeoverPending)
	case "handover_complete":
		m.AcknowledgeRecovery()
	case "fault":
		m.setState(FailoverFault)
	}
}
