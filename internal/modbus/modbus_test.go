package modbus

import (
	"testing"
	"time"
)

func TestInt32RegisterConversion(t *testing.T) {
	tests := []struct {
		name  string
		value int32
	}{
		{"positive", 123456},
		{"negative", -54321},
		{"zero", 0},
		{"max_int32", 2147483647},
		{"min_int32", -2147483648},
		{"small_positive", 1},
		{"small_negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hi, lo := Int32ToRegisters(tt.value)
			got := Int32FromRegisters(hi, lo)
			if got != tt.value {
				t.Errorf("roundtrip failed: input=%d, hi=0x%04x, lo=0x%04x, got=%d",
					tt.value, hi, lo, got)
			}
		})
	}
}

func TestServerClientRoundtrip(t *testing.T) {
	srv := NewServer(":0")
	srv.AllocateRegisters(1, 20)

	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	defer srv.Stop()

	client := NewClient(srv.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer client.Close()

	t.Run("read_default_zeros", func(t *testing.T) {
		regs, err := client.ReadHoldingRegisters(1, 0, 5)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		for i, v := range regs {
			if v != 0 {
				t.Errorf("register %d: got %d, want 0", i, v)
			}
		}
	})

	t.Run("write_and_read_back", func(t *testing.T) {
		err := client.WriteMultipleRegisters(1, 0, []uint16{100, 200, 300})
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		regs, err := client.ReadHoldingRegisters(1, 0, 3)
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		expected := []uint16{100, 200, 300}
		for i, want := range expected {
			if regs[i] != want {
				t.Errorf("register %d: got %d, want %d", i, regs[i], want)
			}
		}
	})

	t.Run("write_int32_and_read_back", func(t *testing.T) {
		err := client.WriteInt32(1, 10, -99999)
		if err != nil {
			t.Fatalf("write int32: %v", err)
		}

		got, err := client.ReadInt32(1, 10)
		if err != nil {
			t.Fatalf("read int32: %v", err)
		}

		if got != -99999 {
			t.Errorf("int32 roundtrip: got %d, want -99999", got)
		}
	})

	t.Run("server_set_client_read", func(t *testing.T) {
		srv.SetRegister(1, 15, 42)
		regs, err := client.ReadHoldingRegisters(1, 15, 1)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if regs[0] != 42 {
			t.Errorf("got %d, want 42", regs[0])
		}
	})

	t.Run("read_invalid_unit_returns_error", func(t *testing.T) {
		_, err := client.ReadHoldingRegisters(99, 0, 1)
		if err == nil {
			t.Error("expected error for invalid unit ID")
		}
	})
}
