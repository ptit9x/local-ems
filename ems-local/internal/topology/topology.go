package topology

import (
	"fmt"
	"sync"

	"github.com/vmo/ems-platform/ems-local/internal/config"
)

// Topology represents the site device hierarchy.
type Topology struct {
	SiteID     string      `json:"site_id" yaml:"site_id"`
	PCCMeter   string      `json:"pcc_meter" yaml:"pcc_meter"`
	SolarMeter string      `json:"solar_meter,omitempty" yaml:"solar_meter,omitempty"`
	LoadMeter  string      `json:"load_meter,omitempty" yaml:"load_meter,omitempty"`
	BESSGroups []BESSGroup `json:"bess_groups" yaml:"bess_groups"`
	EVClusters []EVCluster `json:"ev_clusters,omitempty" yaml:"ev_clusters,omitempty"`
}

type BESSGroup struct {
	Name        string   `json:"name" yaml:"name"`
	BMS         []string `json:"bms" yaml:"bms"`
	PCS         []string `json:"pcs" yaml:"pcs"`
	CommandMode string   `json:"command_mode" yaml:"command_mode"`
}

type EVCluster struct {
	Name     string   `json:"name" yaml:"name"`
	Chargers []string `json:"chargers" yaml:"chargers"`
}

type DeviceInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`                   // "BMS", "PCS", "Meter", "EVCharger"
	Role string `json:"role,omitempty"`         // for meters: "grid", "solar", "load"
}

type Manager struct {
	mu       sync.RWMutex
	topology Topology
	devices  map[string]DeviceInfo // all registered device IDs → type info
}

func NewManager(topo Topology) *Manager {
	return &Manager{
		topology: topo,
		devices:  make(map[string]DeviceInfo),
	}
}

func (m *Manager) Get() Topology {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topology
}

func (m *Manager) Update(topo Topology) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldTopo := m.topology
	m.topology = topo
	if err := m.validateNoLock(); err != nil {
		m.topology = oldTopo
		return err
	}
	return nil
}

func (m *Manager) RegisterDevice(info DeviceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[info.ID] = info
}

func (m *Manager) GetPCCMeterID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topology.PCCMeter
}

func (m *Manager) GetSolarMeterID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.topology.SolarMeter
}

func (m *Manager) GetBESSGroups() []BESSGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	groups := make([]BESSGroup, len(m.topology.BESSGroups))
	copy(groups, m.topology.BESSGroups)
	return groups
}

func (m *Manager) GetEVClusters() []EVCluster {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clusters := make([]EVCluster, len(m.topology.EVClusters))
	copy(clusters, m.topology.EVClusters)
	return clusters
}

func (m *Manager) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.validateNoLock()
}

func (m *Manager) validateNoLock() error {
	if m.topology.PCCMeter == "" {
		return fmt.Errorf("pcc_meter is required")
	}
	if info, ok := m.devices[m.topology.PCCMeter]; !ok {
		return fmt.Errorf("pcc_meter %s not registered", m.topology.PCCMeter)
	} else if info.Type != "Meter" {
		return fmt.Errorf("pcc_meter %s must be a Meter", m.topology.PCCMeter)
	}

	if m.topology.SolarMeter != "" {
		if info, ok := m.devices[m.topology.SolarMeter]; !ok {
			return fmt.Errorf("solar_meter %s not registered", m.topology.SolarMeter)
		} else if info.Type != "Meter" {
			return fmt.Errorf("solar_meter %s must be a Meter", m.topology.SolarMeter)
		}
	}

	if m.topology.LoadMeter != "" {
		if info, ok := m.devices[m.topology.LoadMeter]; !ok {
			return fmt.Errorf("load_meter %s not registered", m.topology.LoadMeter)
		} else if info.Type != "Meter" {
			return fmt.Errorf("load_meter %s must be a Meter", m.topology.LoadMeter)
		}
	}

	for _, g := range m.topology.BESSGroups {
		for _, bms := range g.BMS {
			if info, ok := m.devices[bms]; !ok {
				return fmt.Errorf("bms %s in group %s not registered", bms, g.Name)
			} else if info.Type != "BMS" {
				return fmt.Errorf("device %s in group %s must be a BMS", bms, g.Name)
			}
		}
		for _, pcs := range g.PCS {
			if info, ok := m.devices[pcs]; !ok {
				return fmt.Errorf("pcs %s in group %s not registered", pcs, g.Name)
			} else if info.Type != "PCS" {
				return fmt.Errorf("device %s in group %s must be a PCS", pcs, g.Name)
			}
		}
	}

	for _, cl := range m.topology.EVClusters {
		for _, ev := range cl.Chargers {
			if info, ok := m.devices[ev]; !ok {
				return fmt.Errorf("ev charger %s in cluster %s not registered", ev, cl.Name)
			} else if info.Type != "EVCharger" {
				return fmt.Errorf("device %s in cluster %s must be an EVCharger", ev, cl.Name)
			}
		}
	}
	return nil
}

func BuildFromConfig(cfg config.TopologyConfig, devices config.DevicesConfig) (*Manager, error) {
	topo := Topology{
		PCCMeter:   cfg.PCCMeter,
		SolarMeter: cfg.SolarMeter,
		LoadMeter:  cfg.LoadMeter,
		BESSGroups: make([]BESSGroup, len(cfg.BESSGroups)),
	}

	for i, g := range cfg.BESSGroups {
		topo.BESSGroups[i] = BESSGroup{
			Name:        g.Name,
			BMS:         g.BMS,
			PCS:         g.PCS,
			CommandMode: g.CommandMode,
		}
	}

	mgr := NewManager(topo)

	for _, d := range devices.BMS {
		mgr.RegisterDevice(DeviceInfo{ID: d.ID, Type: "BMS"})
	}
	for _, d := range devices.PCS {
		mgr.RegisterDevice(DeviceInfo{ID: d.ID, Type: "PCS"})
	}
	for _, d := range devices.Meter {
		mgr.RegisterDevice(DeviceInfo{ID: d.ID, Type: "Meter", Role: d.Role})
	}
	
	evChargers := []string{}
	for _, d := range devices.EVCharger {
		mgr.RegisterDevice(DeviceInfo{ID: d.ID, Type: "EVCharger"})
		evChargers = append(evChargers, d.ID)
	}

	if len(evChargers) > 0 {
		mgr.topology.EVClusters = []EVCluster{
			{
				Name:     "Default Cluster",
				Chargers: evChargers,
			},
		}
	}

	if err := mgr.Validate(); err != nil {
		return nil, err
	}

	return mgr, nil
}
