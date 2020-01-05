package kurobako

import (
	"encoding/json"
	"fmt"
)

// Capabilities of a solver.
type Capabilities int64

const (
	// UniformContinuous indicates that the solver can handle numerical parameters that have uniform continuous range.
	UniformContinuous Capabilities = 1 << iota

	// UniformDiscrete indicates that the solver can handle numerical parameters that have uniform discrete range.
	UniformDiscrete

	// LogUniformContinuous indicates that the solver can handle numerical parameters that have log-uniform continuous range.
	LogUniformContinuous

	// LogUniformDiscrete indicates that the solver can handle numerical parameters that have log-uniform discrete range.
	LogUniformDiscrete

	// Categorical indicates that the solver can handle categorical parameters.
	Categorical

	// Conditional indicates that the solver can handle conditional parameters.
	Conditional

	// MultiObjective indicates that the solver supports multi-objective optimization.
	MultiObjective

	// Concurrent indicates that the solver supports concurrent invocations of the ask method.
	Concurrent

	// AllCapabilities represents all of the capabilities.
	AllCapabilities Capabilities = UniformContinuous | UniformDiscrete | LogUniformContinuous |
		LogUniformDiscrete | Categorical | Conditional | MultiObjective | Concurrent
)

// MarshalJSON encodes a Capabilities value to JSON bytes.
func (r Capabilities) MarshalJSON() ([]byte, error) {
	var xs []string
	if (r & UniformContinuous) != 0 {
		xs = append(xs, "UNIFORM_CONTINUOUS")
	}
	if (r & UniformDiscrete) != 0 {
		xs = append(xs, "UNIFORM_DISCRETE")
	}
	if (r & LogUniformContinuous) != 0 {
		xs = append(xs, "LOG_UNIFORM_CONTINUOUS")
	}
	if (r & LogUniformDiscrete) != 0 {
		xs = append(xs, "LOG_UNIFORM_DISCRETE")
	}
	if (r & Categorical) != 0 {
		xs = append(xs, "CATEGORICAL")
	}
	if (r & Conditional) != 0 {
		xs = append(xs, "CONDITIONAL")
	}
	if (r & MultiObjective) != 0 {
		xs = append(xs, "MULTI_OBJECTIVE")
	}
	if (r & Concurrent) != 0 {
		xs = append(xs, "CONCURRENT")
	}
	return json.Marshal(xs)
}

// UnmarshalJSON decodes a Capabilities value from JSON bytes.
func (r *Capabilities) UnmarshalJSON(data []byte) error {
	var xs []string
	if err := json.Unmarshal(data, &xs); err != nil {
		return err
	}

	*r = 0
	for _, s := range xs {
		switch s {
		case "UNIFORM_CONTINUOUS":
			*r |= UniformContinuous
		case "UNIFORM_DISCRETE":
			*r |= UniformDiscrete
		case "LOG_UNIFORM_CONTINUOUS":
			*r |= LogUniformContinuous
		case "LOG_UNIFORM_DISCRETE":
			*r |= LogUniformDiscrete
		case "CATEGORICAL":
			*r |= Categorical
		case "CONDITIONAL":
			*r |= Conditional
		case "MULTI_OBJECTIVE":
			*r |= MultiObjective
		case "CONCURRENT":
			*r |= Concurrent
		default:
			return fmt.Errorf("unknown `Capability`: %s", s)
		}
	}

	return nil
}
