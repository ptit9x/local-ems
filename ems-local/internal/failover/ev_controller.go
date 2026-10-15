package failover

import "sync"

type EVSafeConfig struct {
	MaxPowerPerChargerW int  `json:"max_power_per_charger_w" yaml:"max_power_per_charger_w"`
	AllowNewSessions    bool `json:"allow_new_sessions" yaml:"allow_new_sessions"`
}

type FailoverEVController struct {
	failover *FailoverManager
	safeMode EVSafeConfig
	mu       sync.RWMutex
}

func NewFailoverEVController(fm *FailoverManager, safeCfg EVSafeConfig) *FailoverEVController {
	return &FailoverEVController{
		failover: fm,
		safeMode: safeCfg,
	}
}

func (c *FailoverEVController) ShouldControl() bool {
	state := c.failover.State()
	return state == FailoverNormal || state == FailoverRecovery
}

func (c *FailoverEVController) SafeConfig() EVSafeConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.safeMode
}

func (c *FailoverEVController) IsFailoverActive() bool {
	return c.failover.State() == FailoverActive
}
