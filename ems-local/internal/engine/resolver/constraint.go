package resolver

// Constraint represents a power constraint set by a controller.
type Constraint struct {
	MaxDischargePower *int   // Max discharge in watts (positive), nil = no limit
	MaxChargePower    *int   // Max charge in watts (positive), nil = no limit
	ForcePower        *int   // Force specific power setpoint (negative=charge, positive=discharge)
	Source            string // Controller name that created this constraint
}

// ConstraintCollector collects power constraints from various controllers.
type ConstraintCollector struct {
	constraints []Constraint
}

// NewConstraintCollector creates a new ConstraintCollector.
func NewConstraintCollector() *ConstraintCollector {
	return &ConstraintCollector{
		constraints: make([]Constraint, 0),
	}
}

// Add adds a new constraint to the collector.
func (c *ConstraintCollector) Add(constraint Constraint) {
	c.constraints = append(c.constraints, constraint)
}

// Reset clears all collected constraints.
func (c *ConstraintCollector) Reset() {
	c.constraints = c.constraints[:0]
}

// Constraints returns all collected constraints.
func (c *ConstraintCollector) Constraints() []Constraint {
	return c.constraints
}
