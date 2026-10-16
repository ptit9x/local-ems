package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// RemoteCommand represents a command sent from Cloud to Local EMS.
// Schema: contracts/commands.schema.json
type RemoteCommand struct {
	SchemaVersion int                    `json:"schema_version"`
	SiteID        string                 `json:"site_id"`
	Command       string                 `json:"command"`
	RequestID     string                 `json:"request_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
}

// CommandHandler processes a remote command and returns a response payload and error.
type CommandHandler func(cmd RemoteCommand) error

// CommandSubscriber listens for remote commands from the cloud via MQTT.
type CommandSubscriber struct {
	publisher   *MQTTPublisher
	siteID      string
	topicPrefix string
	handlers    map[string]CommandHandler
	mu          sync.Mutex
	seen        map[string]time.Time
}

// NewCommandSubscriber creates a new CommandSubscriber.
func NewCommandSubscriber(pub *MQTTPublisher, siteID, topicPrefix string) *CommandSubscriber {
	return &CommandSubscriber{
		publisher:   pub,
		siteID:      siteID,
		topicPrefix: topicPrefix,
		handlers:    make(map[string]CommandHandler),
		seen:        make(map[string]time.Time),
	}
}

// OnCommand registers a handler for a specific command type.
func (s *CommandSubscriber) OnCommand(commandType string, handler CommandHandler) {
	s.handlers[commandType] = handler
}

// Start subscribes to the command topic and begins dispatching.
func (s *CommandSubscriber) Start() error {
	topic := fmt.Sprintf("%s/%s/command", s.topicPrefix, s.siteID)

	return s.publisher.Subscribe(topic, func(t string, payload []byte) {
		var cmd RemoteCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			log.Printf("command: failed to parse command: %v (payload=%s)", err, string(payload))
			return
		}
		if cmd.SchemaVersion != 1 || cmd.SiteID != s.siteID || cmd.RequestID == "" || cmd.Command == "" {
			s.publishResponse(cmd.RequestID, "error", map[string]string{"error": "invalid command contract"})
			return
		}
		if cmd.Timestamp.Before(time.Now().Add(-5*time.Minute)) || cmd.Timestamp.After(time.Now().Add(time.Minute)) {
			s.publishResponse(cmd.RequestID, "error", map[string]string{"error": "stale command"})
			return
		}
		s.mu.Lock()
		if _, duplicate := s.seen[cmd.RequestID]; duplicate {
			s.mu.Unlock()
			s.publishResponse(cmd.RequestID, "ok", map[string]string{"deduplicated": "true"})
			return
		}
		s.seen[cmd.RequestID] = time.Now()
		for id, seenAt := range s.seen {
			if time.Since(seenAt) > 24*time.Hour {
				delete(s.seen, id)
			}
		}
		s.mu.Unlock()

		log.Printf("command: received [%s] request_id=%s", cmd.Command, cmd.RequestID)

		handler, ok := s.handlers[cmd.Command]
		if !ok {
			log.Printf("command: unknown command type: %s", cmd.Command)
			s.publishResponse(cmd.RequestID, "error", map[string]string{
				"error": fmt.Sprintf("unknown command: %s", cmd.Command),
			})
			return
		}

		if err := handler(cmd); err != nil {
			log.Printf("command: handler error for [%s]: %v", cmd.Command, err)
			s.publishResponse(cmd.RequestID, "error", map[string]string{
				"error": err.Error(),
			})
			return
		}

		log.Printf("command: [%s] completed successfully", cmd.Command)
		s.publishResponse(cmd.RequestID, "ok", nil)
	})
}

// CommandResponse is sent back to the cloud after processing a command.
type CommandResponse struct {
	SchemaVersion int         `json:"schema_version"`
	RequestID     string      `json:"request_id"`
	SiteID        string      `json:"site_id"`
	Status        string      `json:"status"` // "ok" or "error"
	Timestamp     time.Time   `json:"timestamp"`
	Result        interface{} `json:"result,omitempty"`
}

func (s *CommandSubscriber) publishResponse(requestID, status string, result interface{}) {
	resp := CommandResponse{
		SchemaVersion: 1,
		RequestID:     requestID,
		SiteID:        s.siteID,
		Status:        status,
		Timestamp:     time.Now().UTC(),
		Result:        result,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		log.Printf("command: failed to marshal response: %v", err)
		return
	}

	topic := fmt.Sprintf("%s/%s/command/ack", s.topicPrefix, s.siteID)
	if err := s.publisher.Publish(topic, payload); err != nil {
		log.Printf("command: failed to publish response: %v", err)
	}
}
