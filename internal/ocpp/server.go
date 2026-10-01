package ocpp

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CentralSystem struct {
	mu          sync.RWMutex
	connections map[string]*ChargePoint // chargePointId → connection
	handler     MessageHandler
	server      *http.Server
}

type ChargePoint struct {
	ID              string
	Vendor          string
	Model           string
	Status          string // "Available", "Charging", etc.
	ConnectorStatus map[int]string
	LastHeartbeat   time.Time
	conn            net.Conn
	mu              sync.Mutex
	pending         map[string]chan []byte // uniqueID → response channel
	txCounter       int
}

type MessageHandler interface {
	OnBootNotification(cpID string, req BootNotificationReq) BootNotificationConf
	OnHeartbeat(cpID string) HeartbeatConf
	OnStatusNotification(cpID string, req StatusNotificationReq) StatusNotificationConf
	OnMeterValues(cpID string, req MeterValuesReq) MeterValuesConf
	OnStartTransaction(cpID string, req StartTransactionReq) StartTransactionConf
	OnStopTransaction(cpID string, req StopTransactionReq) StopTransactionConf
}

func NewCentralSystem(addr string, handler MessageHandler) *CentralSystem {
	cs := &CentralSystem{
		connections: make(map[string]*ChargePoint),
		handler:     handler,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ocpp/", cs.handleWebSocket)

	cs.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return cs
}

func (cs *CentralSystem) Start() error {
	go func() {
		if err := cs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("CentralSystem server error: %v", err)
		}
	}()
	return nil
}

func (cs *CentralSystem) Stop() error {
	return cs.server.Close()
}

func (cs *CentralSystem) GetChargePoint(id string) *ChargePoint {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.connections[id]
}

func (cs *CentralSystem) ListChargePoints() []*ChargePoint {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	var list []*ChargePoint
	for _, cp := range cs.connections {
		list = append(list, cp)
	}
	return list
}

func (cs *CentralSystem) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract chargePointId from URL path /ocpp/{chargePointId}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "missing charge point id", http.StatusBadRequest)
		return
	}
	cpID := parts[2]

	conn, err := upgradeWebSocket(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed for %s: %v", cpID, err)
		return
	}

	cp := &ChargePoint{
		ID:              cpID,
		Status:          "Available",
		ConnectorStatus: make(map[int]string),
		LastHeartbeat:   time.Now(),
		conn:            conn,
		pending:         make(map[string]chan []byte),
	}

	cs.mu.Lock()
	if old, exists := cs.connections[cpID]; exists {
		old.conn.Close() // Disconnect old if any
	}
	cs.connections[cpID] = cp
	cs.mu.Unlock()

	defer func() {
		conn.Close()
		cs.mu.Lock()
		if cs.connections[cpID] == cp {
			delete(cs.connections, cpID)
		}
		cs.mu.Unlock()
	}()

	cp.readLoop(cs)
}

func (cp *ChargePoint) readLoop(cs *CentralSystem) {
	for {
		payload, err := readWSFrame(cp.conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("ChargePoint %s error reading ws frame: %v", cp.ID, err)
			}
			return
		}
		
		var msg []json.RawMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("ChargePoint %s sent invalid JSON: %v", cp.ID, err)
			continue
		}

		if len(msg) < 3 {
			continue
		}

		var msgType int
		json.Unmarshal(msg[0], &msgType)

		var msgID string
		json.Unmarshal(msg[1], &msgID)

		if msgType == MessageTypeCallResult || msgType == MessageTypeCallError {
			cp.mu.Lock()
			ch, ok := cp.pending[msgID]
			cp.mu.Unlock()
			if ok {
				ch <- payload
			}
			continue
		}

		if msgType == MessageTypeCall {
			var action string
			json.Unmarshal(msg[2], &action)

			if len(msg) < 4 {
				continue
			}

			responsePayload := cs.handleAction(cp, action, msg[3])
			
			// Send response
			res := []interface{}{MessageTypeCallResult, msgID, responsePayload}
			resBytes, _ := json.Marshal(res)
			writeWSFrame(cp.conn, resBytes)
		}
	}
}

func (cs *CentralSystem) handleAction(cp *ChargePoint, action string, payload json.RawMessage) interface{} {
	switch action {
	case "BootNotification":
		var req BootNotificationReq
		json.Unmarshal(payload, &req)
		cp.Vendor = req.ChargePointVendor
		cp.Model = req.ChargePointModel
		cp.LastHeartbeat = time.Now()
		if cs.handler != nil {
			return cs.handler.OnBootNotification(cp.ID, req)
		}
		return BootNotificationConf{CurrentTime: time.Now().Format(time.RFC3339), Interval: 300, Status: "Accepted"}
	
	case "Heartbeat":
		cp.LastHeartbeat = time.Now()
		if cs.handler != nil {
			return cs.handler.OnHeartbeat(cp.ID)
		}
		return HeartbeatConf{CurrentTime: time.Now().Format(time.RFC3339)}

	case "StatusNotification":
		var req StatusNotificationReq
		json.Unmarshal(payload, &req)
		cp.ConnectorStatus[req.ConnectorId] = req.Status
		if req.ConnectorId == 0 {
			cp.Status = req.Status
		}
		if cs.handler != nil {
			return cs.handler.OnStatusNotification(cp.ID, req)
		}
		return StatusNotificationConf{}

	case "MeterValues":
		var req MeterValuesReq
		json.Unmarshal(payload, &req)
		if cs.handler != nil {
			return cs.handler.OnMeterValues(cp.ID, req)
		}
		return MeterValuesConf{}

	case "StartTransaction":
		var req StartTransactionReq
		json.Unmarshal(payload, &req)
		if cs.handler != nil {
			return cs.handler.OnStartTransaction(cp.ID, req)
		}
		cp.txCounter++
		return StartTransactionConf{TransactionId: cp.txCounter, IdTagInfo: IdTagInfo{Status: "Accepted"}}

	case "StopTransaction":
		var req StopTransactionReq
		json.Unmarshal(payload, &req)
		if cs.handler != nil {
			return cs.handler.OnStopTransaction(cp.ID, req)
		}
		return StopTransactionConf{IdTagInfo: &IdTagInfo{Status: "Accepted"}}
	}

	return struct{}{}
}

func upgradeWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		return nil, fmt.Errorf("not a websocket request")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("hijacking not supported")
	}
	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	bufrw.WriteString("Upgrade: websocket\r\n")
	bufrw.WriteString("Connection: Upgrade\r\n")
	bufrw.WriteString("Sec-WebSocket-Accept: " + acceptKey + "\r\n")
	if r.Header.Get("Sec-WebSocket-Protocol") != "" {
		bufrw.WriteString("Sec-WebSocket-Protocol: ocpp1.6\r\n")
	}
	bufrw.WriteString("\r\n")
	bufrw.Flush()

	return conn, nil
}

// maxOCPPFrameSize limits WebSocket frame payloads to prevent OOM from oversized frames.
const maxOCPPFrameSize = 1 << 20 // 1 MB — more than enough for any OCPP message

func readWSFrame(conn net.Conn) ([]byte, error) {
	var header [2]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return nil, err
	}
	opcode := header[0] & 0x0f
	if opcode == 8 {
		return nil, io.EOF
	}
	mask := header[1]&0x80 != 0
	if !mask {
		return nil, fmt.Errorf("client frame must be masked")
	}
	length := int(header[1] & 0x7f)
	if length == 126 {
		var ext [2]byte
		if _, err := io.ReadFull(conn, ext[:]); err != nil {
			return nil, err
		}
		length = int(binary.BigEndian.Uint16(ext[:]))
	} else if length == 127 {
		var ext [8]byte
		if _, err := io.ReadFull(conn, ext[:]); err != nil {
			return nil, err
		}
		length = int(binary.BigEndian.Uint64(ext[:]))
	}
	if length > maxOCPPFrameSize {
		return nil, fmt.Errorf("frame too large: %d bytes (max %d)", length, maxOCPPFrameSize)
	}
	var maskKey [4]byte
	if _, err := io.ReadFull(conn, maskKey[:]); err != nil {
		return nil, err
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}
	for i := 0; i < length; i++ {
		payload[i] ^= maskKey[i%4]
	}
	return payload, nil
}

func writeWSFrame(conn net.Conn, payload []byte) error {
	header := []byte{0x81} // FIN + Text
	length := len(payload)
	if length < 126 {
		header = append(header, byte(length))
	} else if length <= 65535 {
		header = append(header, 126)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	} else {
		header = append(header, 127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	return nil
}
