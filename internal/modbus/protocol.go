// Package modbus implements a minimal Modbus TCP client and server.
// Supports Function Code 3 (Read Holding Registers) and 16 (Write Multiple Registers).
// Zero external dependencies — pure Go implementation.
package modbus

import (
	"encoding/binary"
	"fmt"
)

// Modbus TCP header size: TransactionID(2) + ProtocolID(2) + Length(2) + UnitID(1) = 7
const mbapHeaderSize = 7

// Function codes
const (
	FuncReadHoldingRegisters  byte = 0x03
	FuncWriteSingleRegister   byte = 0x06
	FuncWriteMultipleRegisters byte = 0x10
)

// Exception codes
const (
	ExcIllegalFunction    byte = 0x01
	ExcIllegalDataAddress byte = 0x02
	ExcIllegalDataValue   byte = 0x03
	ExcServerDeviceFailure byte = 0x04
)

// Frame represents a Modbus TCP frame (MBAP header + PDU).
type Frame struct {
	TransactionID uint16
	ProtocolID    uint16
	UnitID        byte
	FunctionCode  byte
	Data          []byte
}

// MarshalBinary encodes a Frame into wire format.
func (f *Frame) MarshalBinary() ([]byte, error) {
	pduLen := 1 + len(f.Data) // function code + data
	totalLen := mbapHeaderSize + pduLen

	buf := make([]byte, totalLen)
	binary.BigEndian.PutUint16(buf[0:2], f.TransactionID)
	binary.BigEndian.PutUint16(buf[2:4], f.ProtocolID)
	binary.BigEndian.PutUint16(buf[4:6], uint16(pduLen+1)) // length includes UnitID
	buf[6] = f.UnitID
	buf[7] = f.FunctionCode
	copy(buf[8:], f.Data)

	return buf, nil
}

// UnmarshalBinary decodes wire format into a Frame.
func (f *Frame) UnmarshalBinary(data []byte) error {
	if len(data) < mbapHeaderSize+1 {
		return fmt.Errorf("frame too short: %d bytes", len(data))
	}

	f.TransactionID = binary.BigEndian.Uint16(data[0:2])
	f.ProtocolID = binary.BigEndian.Uint16(data[2:4])
	// length at data[4:6] includes UnitID + FunctionCode + Data
	f.UnitID = data[6]
	f.FunctionCode = data[7]

	if len(data) > 8 {
		f.Data = make([]byte, len(data)-8)
		copy(f.Data, data[8:])
	}

	return nil
}

// BuildReadHoldingRegistersRequest creates a FC03 request.
func BuildReadHoldingRegistersRequest(transID uint16, unitID byte, startAddr, quantity uint16) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], transID)
	binary.BigEndian.PutUint16(buf[2:4], 0)       // protocol ID
	binary.BigEndian.PutUint16(buf[4:6], 6)        // length: UnitID(1) + FC(1) + Addr(2) + Qty(2)
	buf[6] = unitID
	buf[7] = FuncReadHoldingRegisters
	binary.BigEndian.PutUint16(buf[8:10], startAddr)
	binary.BigEndian.PutUint16(buf[10:12], quantity)
	return buf
}

// BuildWriteMultipleRegistersRequest creates a FC16 request.
func BuildWriteMultipleRegistersRequest(transID uint16, unitID byte, startAddr uint16, values []uint16) []byte {
	qty := uint16(len(values))
	byteCount := byte(qty * 2)
	// MBAP header(7) + FC(1) + StartAddr(2) + Qty(2) + ByteCount(1) + Data(qty*2)
	totalLen := 7 + 1 + 2 + 2 + 1 + int(byteCount)
	buf := make([]byte, totalLen)

	binary.BigEndian.PutUint16(buf[0:2], transID)
	binary.BigEndian.PutUint16(buf[2:4], 0) // protocol ID
	binary.BigEndian.PutUint16(buf[4:6], uint16(totalLen-6)) // length field
	buf[6] = unitID
	buf[7] = FuncWriteMultipleRegisters
	binary.BigEndian.PutUint16(buf[8:10], startAddr)
	binary.BigEndian.PutUint16(buf[10:12], qty)
	buf[12] = byteCount

	for i, v := range values {
		binary.BigEndian.PutUint16(buf[13+i*2:15+i*2], v)
	}

	return buf
}

// ParseReadHoldingRegistersResponse parses a FC03 response and returns register values.
func ParseReadHoldingRegistersResponse(data []byte) ([]uint16, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	fc := data[7]
	if fc&0x80 != 0 {
		excCode := byte(0)
		if len(data) > 8 {
			excCode = data[8]
		}
		return nil, fmt.Errorf("modbus exception: FC=0x%02x, code=0x%02x", fc, excCode)
	}

	if fc != FuncReadHoldingRegisters {
		return nil, fmt.Errorf("unexpected function code: 0x%02x", fc)
	}

	byteCount := int(data[8])
	if len(data) < 9+byteCount {
		return nil, fmt.Errorf("data truncated: expected %d bytes, got %d", 9+byteCount, len(data))
	}

	regCount := byteCount / 2
	regs := make([]uint16, regCount)
	for i := 0; i < regCount; i++ {
		regs[i] = binary.BigEndian.Uint16(data[9+i*2 : 11+i*2])
	}

	return regs, nil
}

// Int32FromRegisters converts two consecutive uint16 registers to int32 (big-endian).
func Int32FromRegisters(hi, lo uint16) int32 {
	return int32(uint32(hi)<<16 | uint32(lo))
}

// Int32ToRegisters converts int32 to two uint16 registers (big-endian).
func Int32ToRegisters(v int32) (hi, lo uint16) {
	u := uint32(v)
	return uint16(u >> 16), uint16(u & 0xFFFF)
}
