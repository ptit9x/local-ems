package failover

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
)

// HeartbeatMessage is sent every 30s to Cloud.
type HeartbeatMessage struct {
	SiteID    string          `json:"site_id"`
	Timestamp time.Time       `json:"timestamp"`
	Status    string          `json:"status"` // "online", "degraded"
	EVStates  []EVChargerInfo `json:"ev_states"`
	Uptime    int64           `json:"uptime_seconds"`
}

type EVChargerInfo struct {
	ID               string `json:"id"`
	Status           int    `json:"status"`
	ActivePowerW     int    `json:"active_power_w"`
	VehicleConnected bool   `json:"vehicle_connected"`
	Protocol         string `json:"protocol"` // "modbus" or "ocpp"
}

// HeartbeatPublisher sends periodic heartbeats to Cloud.
type HeartbeatPublisher struct {
	publisher   emsSync.Publisher
	siteID      string
	topicPrefix string
	interval    time.Duration
	startTime   time.Time
	evProvider  func() []EVChargerInfo
	done        chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

func NewHeartbeatPublisher(pub emsSync.Publisher, siteID, topicPrefix string, interval time.Duration) *HeartbeatPublisher {
	return &HeartbeatPublisher{
		publisher:   pub,
		siteID:      siteID,
		topicPrefix: topicPrefix,
		interval:    interval,
		startTime:   time.Now(),
		done:        make(chan struct{}),
	}
}

func (h *HeartbeatPublisher) SetEVProvider(fn func() []EVChargerInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.evProvider = fn
}

func (h *HeartbeatPublisher) Start(ctx context.Context) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-h.done:
				return
			case <-ticker.C:
				h.publish()
			}
		}
	}()
}

func (h *HeartbeatPublisher) publish() {
	h.mu.RLock()
	provider := h.evProvider
	h.mu.RUnlock()

	var evStates []EVChargerInfo
	if provider != nil {
		evStates = provider()
	} else {
		evStates = []EVChargerInfo{}
	}

	msg := HeartbeatMessage{
		SiteID:    h.siteID,
		Timestamp: time.Now(),
		Status:    "online",
		EVStates:  evStates,
		Uptime:    int64(time.Since(h.startTime).Seconds()),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("HeartbeatPublisher: error marshaling message: %v\n", err)
		return
	}

	topic := h.topicPrefix + "/" + h.siteID + "/heartbeat"
	_ = h.publisher.Publish(topic, data)
}

func (h *HeartbeatPublisher) Stop() {
	close(h.done)
	h.wg.Wait()
}
