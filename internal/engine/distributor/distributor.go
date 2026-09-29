// Package distributor splits a total power setpoint across multiple PCS units.
package distributor

import (
	"github.com/vmo/local-ems/internal/devices/aggregator"
)

// Strategy defines how to distribute power across PCS units.
type Strategy string

const (
	EqualSplit      Strategy = "equal"        // split evenly across active PCS
	ProportionalSOC Strategy = "proportional" // weight by connected BMS SOC
	RoundRobin      Strategy = "round_robin"  // rotate lead PCS each cycle
)

// BESSGroup defines a group of PCS units with a command mode.
type BESSGroup struct {
	Name        string
	PCSIndices  []int    // indices into the PCS array
	CommandMode string   // "master" or "individual"
	MasterIndex int      // which PCS index is the master (first in group)
}

// Distributor splits a total setpoint across N PCS units.
type Distributor struct {
	strategy    Strategy
	pcsCount    int
	rrIndex     int  // for round-robin: which PCS leads
	groups      []BESSGroup
}


// New creates a Distributor with the given strategy and PCS count.
func New(strategy Strategy, pcsCount int) *Distributor {
	return &Distributor{
		strategy: strategy,
		pcsCount: pcsCount,
	}
}

// SetGroups configures the distributor with BESS groups for master/individual modes.
func (d *Distributor) SetGroups(groups []BESSGroup) {
	d.groups = groups
}

// Result is the per-PCS setpoint output.
type Result struct {
	Setpoints []int  // setpoint for each PCS (index-aligned)
	TotalW    int    // sum of all setpoints (may differ from input due to clamping)
	Clamped   bool   // true if any PCS was clamped to its max
}

// Distribute splits totalW across PCS units based on their status.
// Positive totalW = discharge, negative = charge.
func (d *Distributor) Distribute(totalW int, pcsStates []aggregator.PCSState) Result {
	n := len(pcsStates)
	if n == 0 {
		return Result{Setpoints: nil}
	}

	var setpoints []int

	if len(d.groups) > 0 {
		setpoints = d.distributeWithGroups(totalW, pcsStates)
	} else {
		switch d.strategy {
		case EqualSplit:
			setpoints = d.equalSplit(totalW, pcsStates)
		case RoundRobin:
			setpoints = d.roundRobin(totalW, pcsStates)
		default:
			setpoints = d.equalSplit(totalW, pcsStates)
		}
	}

	// Calculate result
	result := Result{Setpoints: setpoints}
	for i, sp := range setpoints {
		result.TotalW += sp
		if pcsStates[i].MaxPowerW > 0 && abs(sp) >= pcsStates[i].MaxPowerW {
			result.Clamped = true
		}
	}
	return result
}

func (d *Distributor) distributeWithGroups(totalW int, pcs []aggregator.PCSState) []int {
	n := len(pcs)
	setpoints := make([]int, n)

	activePerGroup := make([]int, len(d.groups))
	totalActive := 0
	for i, g := range d.groups {
		active := 0
		for _, idx := range g.PCSIndices {
			if idx >= 0 && idx < n {
				p := pcs[idx]
				if p.Online && p.Status != aggregator.PCSFault {
					active++
				}
			}
		}
		activePerGroup[i] = active
		totalActive += active
	}

	if totalActive == 0 {
		return setpoints
	}

	perActivePCS := totalW / totalActive
	remainder := totalW - (perActivePCS * totalActive)

	for i, g := range d.groups {
		if activePerGroup[i] == 0 {
			continue
		}

		groupW := perActivePCS * activePerGroup[i]
		if remainder != 0 {
			groupW += remainder
			remainder = 0
		}

		if g.CommandMode == "master" {
			masterIdx := -1
			if g.MasterIndex >= 0 && g.MasterIndex < n {
				p := pcs[g.MasterIndex]
				if p.Online && p.Status != aggregator.PCSFault {
					masterIdx = g.MasterIndex
				}
			}
			if masterIdx == -1 {
				for _, idx := range g.PCSIndices {
					if idx >= 0 && idx < n {
						p := pcs[idx]
						if p.Online && p.Status != aggregator.PCSFault {
							masterIdx = idx
							break
						}
					}
				}
			}

			if masterIdx != -1 {
				sp := groupW
				p := pcs[masterIdx]
				if p.MaxPowerW > 0 {
					if sp > p.MaxPowerW {
						sp = p.MaxPowerW
					} else if sp < -p.MaxPowerW {
						sp = -p.MaxPowerW
					}
				}
				setpoints[masterIdx] = sp
			}
		} else {
			activeInGroup := activePerGroup[i]
			perPCS := groupW / activeInGroup
			groupRem := groupW - (perPCS * activeInGroup)

			assigned := 0
			for _, idx := range g.PCSIndices {
				if idx >= 0 && idx < n {
					p := pcs[idx]
					if p.Online && p.Status != aggregator.PCSFault {
						sp := perPCS
						if assigned == 0 {
							sp += groupRem
						}

						if p.MaxPowerW > 0 {
							if sp > p.MaxPowerW {
								sp = p.MaxPowerW
							} else if sp < -p.MaxPowerW {
								sp = -p.MaxPowerW
							}
						}
						setpoints[idx] = sp
						assigned++
					}
				}
			}
		}
	}

	return setpoints
}

func (d *Distributor) equalSplit(totalW int, pcs []aggregator.PCSState) []int {
	n := len(pcs)
	setpoints := make([]int, n)

	// Count active PCS
	active := 0
	for _, p := range pcs {
		if p.Online && p.Status != aggregator.PCSFault {
			active++
		}
	}
	if active == 0 {
		return setpoints
	}

	perPCS := totalW / active
	remainder := totalW - (perPCS * active) // distribute remainder to first PCS

	assigned := 0
	for i, p := range pcs {
		if !p.Online || p.Status == aggregator.PCSFault {
			continue
		}

		sp := perPCS
		if assigned == 0 {
			sp += remainder
		}

		// Clamp to PCS max rating
		if p.MaxPowerW > 0 {
			if sp > p.MaxPowerW {
				sp = p.MaxPowerW
			} else if sp < -p.MaxPowerW {
				sp = -p.MaxPowerW
			}
		}

		setpoints[i] = sp
		assigned++
	}

	return setpoints
}

func (d *Distributor) roundRobin(totalW int, pcs []aggregator.PCSState) []int {
	n := len(pcs)
	setpoints := make([]int, n)

	// Find active PCS indices
	var activeIdx []int
	for i, p := range pcs {
		if p.Online && p.Status != aggregator.PCSFault {
			activeIdx = append(activeIdx, i)
		}
	}
	if len(activeIdx) == 0 {
		return setpoints
	}

	// Rotate: start from rrIndex, fill each PCS up to max before moving to next
	remaining := totalW
	sign := 1
	if totalW < 0 {
		sign = -1
		remaining = -remaining
	}

	startPos := d.rrIndex % len(activeIdx)
	for j := 0; j < len(activeIdx) && remaining > 0; j++ {
		idx := activeIdx[(startPos+j)%len(activeIdx)]
		maxW := pcs[idx].MaxPowerW
		if maxW <= 0 {
			maxW = remaining // no limit
		}

		assign := remaining
		if assign > maxW {
			assign = maxW
		}

		setpoints[idx] = assign * sign
		remaining -= assign
	}

	d.rrIndex++
	return setpoints
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
