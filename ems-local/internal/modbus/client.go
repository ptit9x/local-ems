package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Client is a Modbus TCP client that connects to a single server.
type Client struct {
	addr    string
	conn    net.Conn
	mu      sync.Mutex
	transID uint16
	timeout time.Duration
}

// NewClient creates a new Modbus TCP client.
func NewClient(addr string, timeout time.Duration) *Client {
	return &Client{
		addr:    addr,
		timeout: timeout,
	}
}

// Connect establishes the TCP connection.
func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return fmt.Errorf("modbus connect %s: %w", c.addr, err)
	}
	c.conn = conn
	return nil
}

// Close closes the TCP connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ReadHoldingRegisters reads quantity registers starting at startAddr (FC03).
func (c *Client) ReadHoldingRegisters(unitID byte, startAddr, quantity uint16) ([]uint16, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transID++
	req := BuildReadHoldingRegistersRequest(c.transID, unitID, startAddr, quantity)

	if c.conn == nil {
		return nil, fmt.Errorf("modbus client not connected")
	}

	if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}

	if _, err := c.conn.Write(req); err != nil {
		return nil, fmt.Errorf("modbus write: %w", err)
	}

	// Read MBAP header first (7 bytes) to get length
	header := make([]byte, mbapHeaderSize)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, fmt.Errorf("modbus read header: %w", err)
	}

	pduLen := int(binary.BigEndian.Uint16(header[4:6])) - 1 // subtract UnitID already in header
	if pduLen <= 0 || pduLen > 255 {
		return nil, fmt.Errorf("invalid PDU length: %d", pduLen)
	}

	pdu := make([]byte, pduLen)
	if _, err := io.ReadFull(c.conn, pdu); err != nil {
		return nil, fmt.Errorf("modbus read pdu: %w", err)
	}

	// Combine header + PDU for parsing
	full := make([]byte, mbapHeaderSize+pduLen)
	copy(full, header)
	copy(full[mbapHeaderSize:], pdu)

	return ParseReadHoldingRegistersResponse(full)
}

// WriteMultipleRegisters writes values to registers starting at startAddr (FC16).
func (c *Client) WriteMultipleRegisters(unitID byte, startAddr uint16, values []uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transID++
	req := BuildWriteMultipleRegistersRequest(c.transID, unitID, startAddr, values)

	if c.conn == nil {
		return fmt.Errorf("modbus client not connected")
	}

	if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}

	if _, err := c.conn.Write(req); err != nil {
		return fmt.Errorf("modbus write: %w", err)
	}

	// Read response
	resp := make([]byte, 12) // FC16 response is always 12 bytes
	if _, err := io.ReadFull(c.conn, resp); err != nil {
		return fmt.Errorf("modbus read response: %w", err)
	}

	fc := resp[7]
	if fc&0x80 != 0 {
		excCode := byte(0)
		if len(resp) > 8 {
			excCode = resp[8]
		}
		return fmt.Errorf("modbus exception: FC=0x%02x, code=0x%02x", fc, excCode)
	}

	return nil
}

// WriteSingleRegisterValues writes a single int32 value as two registers (hi, lo).
func (c *Client) WriteInt32(unitID byte, startAddr uint16, value int32) error {
	hi, lo := Int32ToRegisters(value)
	return c.WriteMultipleRegisters(unitID, startAddr, []uint16{hi, lo})
}

// ReadInt32 reads two consecutive registers and returns as int32.
func (c *Client) ReadInt32(unitID byte, startAddr uint16) (int32, error) {
	regs, err := c.ReadHoldingRegisters(unitID, startAddr, 2)
	if err != nil {
		return 0, err
	}
	if len(regs) < 2 {
		return 0, fmt.Errorf("expected 2 registers, got %d", len(regs))
	}
	return Int32FromRegisters(regs[0], regs[1]), nil
}
