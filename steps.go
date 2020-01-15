package kurobako

import (
	"encoding/json"
	"fmt"
)

// Steps represents a sequence of evaluable steps of a problem.
//
// At each evaluable step, the evaluator of the problem can suspend the evaluation and return intermediate evaluation values at that step.
type Steps struct {
	isSequential bool
	steps        []uint64
}

// NewSteps creates a new Steps instance.
func NewSteps(steps []uint64) (*Steps, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("empty steps isn't allowed")
	}

	isSequential := true
	curr := steps[0]
	for _, step := range steps[1:] {
		if curr >= step {
			return nil, fmt.Errorf("steps should be monotonically increasing")
		}

		isSequential = isSequential && curr+1 == step
		curr = step
	}

	if isSequential {
		return &Steps{true, []uint64{steps[len(steps)-1]}}, nil
	}

	return &Steps{false, steps}, nil
}

// Last returns the last step.
func (r Steps) Last() uint64 {
	return r.steps[len(r.steps)-1]
}

// AsSlice returns the steps as slice.
func (r Steps) AsSlice() []uint64 {
	if r.isSequential {
		last := r.Last()
		steps := make([]uint64, 0, last)
		for i := uint64(1); i <= last; i++ {
			steps = append(steps, i)
		}
		return steps
	}

	return r.steps
}

// MarshalJSON encodes a Steps object to JSON bytes.
func (r Steps) MarshalJSON() ([]byte, error) {
	if r.isSequential {
		return json.Marshal(r.Last())
	}

	return json.Marshal(r.steps)
}

// UnmarshalJSON decodes a Steps object from JSON bytes.
func (r *Steps) UnmarshalJSON(data []byte) error {
	var lastStep uint64
	if json.Unmarshal(data, &lastStep) == nil {
		r.isSequential = true
		r.steps = []uint64{lastStep}
		return nil
	}

	var steps []uint64
	if err := json.Unmarshal(data, &steps); err != nil {
		return err
	}
	r.isSequential = false
	r.steps = steps
	return nil
}
