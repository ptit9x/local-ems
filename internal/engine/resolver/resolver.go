package resolver

// Resolver computes the final ESS power setpoint from collected constraints.
type Resolver struct {
	collector *ConstraintCollector
}

// ResolvedPower is the output of the resolver after evaluating all constraints.
type ResolvedPower struct {
	Setpoint           int      // Final power setpoint (positive=discharge, negative=charge)
	AppliedConstraints []string // Names of controllers whose constraints affected the result
	Clamped            bool     // True if the requested power was modified by constraints
}

// NewResolver creates a new Resolver backed by the given collector.
func NewResolver(collector *ConstraintCollector) *Resolver {
	return &Resolver{collector: collector}
}

// Resolve takes a requested power setpoint and clamps it within the bounds
// defined by all collected constraints.
//
// Convention: positive = discharge, negative = charge.
// MaxDischargePower and MaxChargePower are expressed as positive watts.
func (r *Resolver) Resolve(requestedPower int) ResolvedPower {
	constraints := r.collector.Constraints()

	if len(constraints) == 0 {
		return ResolvedPower{
			Setpoint: requestedPower,
		}
	}

	var (
		maxDischarge *int    // strictest (lowest) max discharge
		maxCharge    *int    // strictest (lowest) max charge
		forcePower   *int    // first (highest-priority) force power
		applied      []string
		clamped      bool
	)

	for _, c := range constraints {
		if c.MaxDischargePower != nil {
			if maxDischarge == nil || *c.MaxDischargePower < *maxDischarge {
				maxDischarge = c.MaxDischargePower
			}
		}
		if c.MaxChargePower != nil {
			if maxCharge == nil || *c.MaxChargePower < *maxCharge {
				maxCharge = c.MaxChargePower
			}
		}
		if c.ForcePower != nil && forcePower == nil {
			// First ForcePower wins (highest priority controller added first)
			forcePower = c.ForcePower
		}
	}

	setpoint := requestedPower

	// ForcePower overrides the requested power
	if forcePower != nil {
		setpoint = *forcePower
		clamped = setpoint != requestedPower
	}

	// Clamp discharge (positive direction)
	if maxDischarge != nil && setpoint > *maxDischarge {
		setpoint = *maxDischarge
		clamped = true
	}

	// Clamp charge (negative direction): maxCharge is positive watts,
	// so the most negative allowed is -maxCharge
	if maxCharge != nil && setpoint < -(*maxCharge) {
		setpoint = -(*maxCharge)
		clamped = true
	}

	// Collect which constraints were applied
	for _, c := range constraints {
		if c.MaxDischargePower != nil || c.MaxChargePower != nil || c.ForcePower != nil {
			applied = append(applied, c.Source)
		}
	}

	return ResolvedPower{
		Setpoint:           setpoint,
		AppliedConstraints: applied,
		Clamped:            clamped,
	}
}
