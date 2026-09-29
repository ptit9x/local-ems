package devices

import (
	"fmt"
	"sync"

	"github.com/vmo/local-ems/internal/modbus"
)

// PCSProfile defines the register map and configuration for a specific PCS brand.
type PCSProfile struct {
	Brand             string
	ActivePowerSetReg uint16
	ActivePowerActReg uint16
	ReactivePowerReg  uint16
	StatusReg         uint16
	RegisterCount     uint16
	ScaleFactor       float64
	StatusMap         map[uint16]PCSStatus
	CommandMode       string
	Description       string
}

// PCSProfileRegistry manages supported PCS profiles.
type PCSProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]PCSProfile
}

// Register adds a new PCSProfile to the registry.
func (r *PCSProfileRegistry) Register(profile PCSProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if profile.Brand == "" {
		return fmt.Errorf("registering profile: brand cannot be empty")
	}
	r.profiles[profile.Brand] = profile
	return nil
}

// Get retrieves a PCSProfile by brand name.
func (r *PCSProfileRegistry) Get(brand string) (PCSProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[brand]
	if !ok {
		return PCSProfile{}, fmt.Errorf("getting profile: brand '%s' not found", brand)
	}
	return p, nil
}

// List returns all registered profile brand names.
func (r *PCSProfileRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []string
	for b := range r.profiles {
		list = append(list, b)
	}
	return list
}

// DefaultProfile returns the generic profile.
func (r *PCSProfileRegistry) DefaultProfile() PCSProfile {
	p, _ := r.Get("generic")
	return p
}

// DefaultPCSProfileRegistry returns a pre-loaded registry with standard profiles.
func DefaultPCSProfileRegistry() *PCSProfileRegistry {
	r := &PCSProfileRegistry{
		profiles: make(map[string]PCSProfile),
	}

	r.Register(PCSProfile{
		Brand:             "generic",
		ActivePowerSetReg: 0,
		ActivePowerActReg: 1,
		ReactivePowerReg:  2,
		StatusReg:         3,
		RegisterCount:     4,
		ScaleFactor:       1.0,
		StatusMap:         make(map[uint16]PCSStatus),
		CommandMode:       "master",
		Description:       "Generic profile with default registers",
	})

	r.Register(PCSProfile{
		Brand:             "sungrow",
		ActivePowerSetReg: 5040,
		ActivePowerActReg: 5008,
		ReactivePowerReg:  5010,
		StatusReg:         5000,
		RegisterCount:     41,
		ScaleFactor:       1.0,
		StatusMap: map[uint16]PCSStatus{
			0: PCSStatus(0), // Standby
			1: PCSStatus(1), // Running
			2: PCSStatus(2), // Fault
			4: PCSStatus(4), // Shutdown
		},
		CommandMode:       "master",
		Description:       "Sungrow PCS profile",
	})

	r.Register(PCSProfile{
		Brand:             "sinexcel",
		ActivePowerSetReg: 100,
		ActivePowerActReg: 102,
		ReactivePowerReg:  104,
		StatusReg:         110,
		RegisterCount:     11,
		ScaleFactor:       10.0,
		StatusMap:         make(map[uint16]PCSStatus),
		CommandMode:       "master",
		Description:       "Sinexcel PCS profile",
	})

	return r
}

// ProfiledPCSAdapter wraps PCSAdapter-like functionality using a PCSProfile.
type ProfiledPCSAdapter struct {
	mu      sync.RWMutex
	client  *modbus.Client
	unitID  byte
	profile PCSProfile
	data    PCSReading
}

// NewProfiledPCSAdapter creates a new adapter with a specific profile.
func NewProfiledPCSAdapter(client *modbus.Client, unitID byte, profile PCSProfile) *ProfiledPCSAdapter {
	return &ProfiledPCSAdapter{
		client:  client,
		unitID:  unitID,
		profile: profile,
	}
}

// Poll reads data from the PCS device using the configured profile's registers.
func (a *ProfiledPCSAdapter) Poll() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Read Status
	statusRegs, err := a.client.ReadHoldingRegisters(a.unitID, a.profile.StatusReg, 1)
	if err != nil {
		return fmt.Errorf("polling status register: %w", err)
	}
	if len(statusRegs) > 0 {
		rawStatus := statusRegs[0]
		if mappedStatus, ok := a.profile.StatusMap[rawStatus]; ok {
			a.data.Status = mappedStatus
		} else {
			a.data.Status = PCSStatus(rawStatus)
		}
	}

	// Read Actual Active Power
	powerRegs, err := a.client.ReadHoldingRegisters(a.unitID, a.profile.ActivePowerActReg, 1)
	if err != nil {
		return fmt.Errorf("polling active power register: %w", err)
	}
	if len(powerRegs) > 0 {
		a.data.ActualActivePowerW = int(float64(int16(powerRegs[0])) * a.profile.ScaleFactor)
	}

	// Read Reactive Power
	reactiveRegs, err := a.client.ReadHoldingRegisters(a.unitID, a.profile.ReactivePowerReg, 1)
	if err != nil {
		return fmt.Errorf("polling reactive power register: %w", err)
	}
	if len(reactiveRegs) > 0 {
		a.data.ReactivePowerVar = int(float64(int16(reactiveRegs[0])) * a.profile.ScaleFactor)
	}

	return nil
}

// Data returns the current PCSReading.
func (a *ProfiledPCSAdapter) Data() PCSReading {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

// WriteActivePower writes a new active power setpoint using the profile.
func (a *ProfiledPCSAdapter) WriteActivePower(watts int) error {
	val := uint16(float64(watts) / a.profile.ScaleFactor)
	err := a.client.WriteMultipleRegisters(a.unitID, a.profile.ActivePowerSetReg, []uint16{val})
	if err != nil {
		return fmt.Errorf("writing active power setpoint: %w", err)
	}
	return nil
}

// Brand returns the brand of the configured profile.
func (a *ProfiledPCSAdapter) Brand() string {
	return a.profile.Brand
}

// CommandMode returns the command mode of the configured profile.
func (a *ProfiledPCSAdapter) CommandMode() string {
	return a.profile.CommandMode
}
