package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Server is a minimal Modbus TCP server that holds registers in memory.
// Used by the simulator to expose fake device data via Modbus protocol.
type Server struct {
	listener  net.Listener
	registers map[byte][]uint16 // unitID → register array
	mu        sync.RWMutex
	done      chan struct{}
	addr      string
}

// NewServer creates a new Modbus TCP server listening on the given address.
// Pass ":0" for auto-assigned port.
func NewServer(addr string) *Server {
	return &Server{
		registers: make(map[byte][]uint16),
		done:      make(chan struct{}),
		addr:      addr,
	}
}

// AllocateRegisters allocates register space for a unit ID.
func (s *Server) AllocateRegisters(unitID byte, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registers[unitID] = make([]uint16, count)
}

// SetRegister sets a single register value.
func (s *Server) SetRegister(unitID byte, addr uint16, value uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if regs, ok := s.registers[unitID]; ok && int(addr) < len(regs) {
		regs[addr] = value
	}
}

// SetInt32 sets an int32 value across two consecutive registers.
func (s *Server) SetInt32(unitID byte, addr uint16, value int32) {
	hi, lo := Int32ToRegisters(value)
	s.mu.Lock()
	defer s.mu.Unlock()
	if regs, ok := s.registers[unitID]; ok && int(addr)+1 < len(regs) {
		regs[addr] = hi
		regs[addr+1] = lo
	}
}

// GetRegister reads a single register value.
func (s *Server) GetRegister(unitID byte, addr uint16) uint16 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if regs, ok := s.registers[unitID]; ok && int(addr) < len(regs) {
		return regs[addr]
	}
	return 0
}

// GetInt32 reads an int32 from two consecutive registers.
func (s *Server) GetInt32(unitID byte, addr uint16) int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if regs, ok := s.registers[unitID]; ok && int(addr)+1 < len(regs) {
		return Int32FromRegisters(regs[addr], regs[addr+1])
	}
	return 0
}

// Start begins listening for connections in the background.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("modbus server listen: %w", err)
	}
	s.listener = ln

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-s.done:
					return
				default:
					log.Printf("modbus server accept error: %v", err)
					continue
				}
			}
			go s.handleConn(conn)
		}
	}()

	return nil
}

// Addr returns the listener address (useful when using ":0" for auto port).
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Stop shuts down the server.
func (s *Server) Stop() error {
	close(s.done)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConn processes a single client connection.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-s.done:
			return
		default:
		}

		// Read MBAP header
		header := make([]byte, mbapHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			return // client disconnected
		}

		pduLen := int(binary.BigEndian.Uint16(header[4:6])) - 1 // subtract UnitID
		if pduLen <= 0 || pduLen > 255 {
			return
		}

		pdu := make([]byte, pduLen)
		if _, err := io.ReadFull(conn, pdu); err != nil {
			return
		}

		unitID := header[6]
		fc := pdu[0]
		transID := binary.BigEndian.Uint16(header[0:2])

		var resp []byte

		switch fc {
		case FuncReadHoldingRegisters:
			resp = s.handleReadHolding(transID, unitID, pdu)
		case FuncWriteMultipleRegisters:
			resp = s.handleWriteMultiple(transID, unitID, pdu)
		case FuncWriteSingleRegister:
			resp = s.handleWriteSingle(transID, unitID, pdu)
		default:
			resp = buildExceptionResponse(transID, unitID, fc, ExcIllegalFunction)
		}

		if _, err := conn.Write(resp); err != nil {
			return
		}
	}
}

// handleReadHolding processes FC03.
func (s *Server) handleReadHolding(transID uint16, unitID byte, pdu []byte) []byte {
	if len(pdu) < 5 {
		return buildExceptionResponse(transID, unitID, FuncReadHoldingRegisters, ExcIllegalDataValue)
	}

	startAddr := binary.BigEndian.Uint16(pdu[1:3])
	quantity := binary.BigEndian.Uint16(pdu[3:5])

	s.mu.RLock()
	regs, ok := s.registers[unitID]
	s.mu.RUnlock()

	if !ok || int(startAddr)+int(quantity) > len(regs) {
		return buildExceptionResponse(transID, unitID, FuncReadHoldingRegisters, ExcIllegalDataAddress)
	}

	byteCount := byte(quantity * 2)
	respLen := 3 + int(byteCount) // UnitID + FC + ByteCount + Data
	resp := make([]byte, 7+int(byteCount)+2)

	binary.BigEndian.PutUint16(resp[0:2], transID)
	binary.BigEndian.PutUint16(resp[2:4], 0)
	binary.BigEndian.PutUint16(resp[4:6], uint16(respLen))
	resp[6] = unitID
	resp[7] = FuncReadHoldingRegisters
	resp[8] = byteCount

	s.mu.RLock()
	for i := uint16(0); i < quantity; i++ {
		binary.BigEndian.PutUint16(resp[9+i*2:11+i*2], regs[startAddr+i])
	}
	s.mu.RUnlock()

	return resp[:9+int(byteCount)]
}

// handleWriteMultiple processes FC16.
func (s *Server) handleWriteMultiple(transID uint16, unitID byte, pdu []byte) []byte {
	if len(pdu) < 6 {
		return buildExceptionResponse(transID, unitID, FuncWriteMultipleRegisters, ExcIllegalDataValue)
	}

	startAddr := binary.BigEndian.Uint16(pdu[1:3])
	quantity := binary.BigEndian.Uint16(pdu[3:5])
	byteCount := pdu[5]

	if len(pdu) < 6+int(byteCount) {
		return buildExceptionResponse(transID, unitID, FuncWriteMultipleRegisters, ExcIllegalDataValue)
	}

	s.mu.Lock()
	regs, ok := s.registers[unitID]
	if !ok || int(startAddr)+int(quantity) > len(regs) {
		s.mu.Unlock()
		return buildExceptionResponse(transID, unitID, FuncWriteMultipleRegisters, ExcIllegalDataAddress)
	}

	for i := uint16(0); i < quantity; i++ {
		regs[startAddr+i] = binary.BigEndian.Uint16(pdu[6+i*2 : 8+i*2])
	}
	s.mu.Unlock()

	// Response: MBAP(7) + FC(1) + StartAddr(2) + Quantity(2) = 12
	resp := make([]byte, 12)
	binary.BigEndian.PutUint16(resp[0:2], transID)
	binary.BigEndian.PutUint16(resp[2:4], 0)
	binary.BigEndian.PutUint16(resp[4:6], 6) // length
	resp[6] = unitID
	resp[7] = FuncWriteMultipleRegisters
	binary.BigEndian.PutUint16(resp[8:10], startAddr)
	binary.BigEndian.PutUint16(resp[10:12], quantity)

	return resp
}

// handleWriteSingle processes FC06.
func (s *Server) handleWriteSingle(transID uint16, unitID byte, pdu []byte) []byte {
	if len(pdu) < 5 {
		return buildExceptionResponse(transID, unitID, FuncWriteSingleRegister, ExcIllegalDataValue)
	}

	addr := binary.BigEndian.Uint16(pdu[1:3])
	value := binary.BigEndian.Uint16(pdu[3:5])

	s.mu.Lock()
	regs, ok := s.registers[unitID]
	if !ok || int(addr) >= len(regs) {
		s.mu.Unlock()
		return buildExceptionResponse(transID, unitID, FuncWriteSingleRegister, ExcIllegalDataAddress)
	}
	regs[addr] = value
	s.mu.Unlock()

	// Echo request as response
	resp := make([]byte, 12)
	binary.BigEndian.PutUint16(resp[0:2], transID)
	binary.BigEndian.PutUint16(resp[2:4], 0)
	binary.BigEndian.PutUint16(resp[4:6], 6)
	resp[6] = unitID
	resp[7] = FuncWriteSingleRegister
	binary.BigEndian.PutUint16(resp[8:10], addr)
	binary.BigEndian.PutUint16(resp[10:12], value)

	return resp
}

// buildExceptionResponse creates a Modbus exception response.
func buildExceptionResponse(transID uint16, unitID, fc, excCode byte) []byte {
	resp := make([]byte, 9)
	binary.BigEndian.PutUint16(resp[0:2], transID)
	binary.BigEndian.PutUint16(resp[2:4], 0)
	binary.BigEndian.PutUint16(resp[4:6], 3) // length
	resp[6] = unitID
	resp[7] = fc | 0x80 // error flag
	resp[8] = excCode
	return resp
}
