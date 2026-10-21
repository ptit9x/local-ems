package device_runtime

import (
	"errors"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/devices/aggregator"
)

type fakeDevice struct {
	id       string
	typ      devices.DeviceType
	data     devices.DataPoints
	writes   int
	failures int
}

func (d *fakeDevice) ID() string                   { return d.id }
func (d *fakeDevice) Type() devices.DeviceType     { return d.typ }
func (d *fakeDevice) Protocol() string             { return "test" }
func (d *fakeDevice) ProtocolType() string         { return "test" }
func (d *fakeDevice) Vendor() string               { return "test" }
func (d *fakeDevice) Status() devices.DeviceStatus { return devices.DeviceStatusOnline }
func (d *fakeDevice) Health() devices.DeviceHealth { return devices.DeviceHealth{Connected: true} }
func (d *fakeDevice) Poll() error                  { return nil }
func (d *fakeDevice) Snapshot() devices.DataPoints { return d.data }
func (d *fakeDevice) Write(devices.Command) error {
	d.writes++
	if d.writes <= d.failures {
		return errors.New("temporary")
	}
	return nil
}

func point(name string, value interface{}) devices.DataPoint {
	return devices.DataPoint{Name: name, Value: value, Timestamp: time.Now(), Quality: "good"}
}

func TestPollAndEVCommandIdempotency(t *testing.T) {
	cfg := config.DefaultDev()
	cfg.Devices.EVCharger = cfg.Devices.EVCharger[:1]
	reg := devices.NewDeviceRegistry()
	ev := &fakeDevice{id: cfg.Devices.EVCharger[0].ID, typ: devices.DeviceTypeEVCharger, data: devices.DataPoints{"status": point("status", 1), "active_power_w": point("active_power_w", 7000), "vehicle_connected": point("vehicle_connected", true)}}
	if err := reg.Register(ev); err != nil {
		t.Fatal(err)
	}
	agg := aggregator.New(0, 0, 0, 1)
	o := New(reg, agg, cfg, func() bool { return true })
	o.SetRetryPolicy(3, 0)
	o.PollOnce()
	if got := agg.Aggregate().TotalEVPowerW; got != 7000 {
		t.Fatalf("power=%d", got)
	}
	ev.failures = 2
	if err := o.ApplyEVLimits(map[string]int{ev.id: 11000}); err != nil {
		t.Fatal(err)
	}
	if ev.writes != 3 {
		t.Fatalf("writes=%d", ev.writes)
	}
	if err := o.ApplyEVLimits(map[string]int{ev.id: 11000}); err != nil {
		t.Fatal(err)
	}
	if ev.writes != 3 {
		t.Fatalf("duplicate command was not suppressed")
	}
}

func TestControlLeaseBlocksCommands(t *testing.T) {
	cfg := config.DefaultDev()
	reg := devices.NewDeviceRegistry()
	ev := &fakeDevice{id: "ev-0", typ: devices.DeviceTypeEVCharger}
	_ = reg.Register(ev)
	o := New(reg, aggregator.New(0, 0, 0, 1), cfg, func() bool { return false })
	if err := o.ApplyEVLimits(map[string]int{"ev-0": 1000}); err == nil {
		t.Fatal("expected lease rejection")
	}
	if ev.writes != 0 {
		t.Fatal("command reached device without lease")
	}
}
