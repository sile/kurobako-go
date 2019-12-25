package kurobako

import (
	"encoding/json"
	"fmt"
)

type Steps struct {
	isSequential bool
	steps        []uint64
}

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
	} else {
		return &Steps{false, steps}, nil
	}
}

func (r Steps) Last() uint64 {
	return r.steps[len(r.steps)-1]
}

func (r Steps) AsSlice() []uint64 {
	if r.isSequential {
		last := r.Last()
		steps := []uint64{}
		for i := uint64(1); i <= last; i++ {
			steps = append(steps, i)
		}
		return steps
	} else {
		return r.steps
	}
}

func (r Steps) MarshalJSON() ([]byte, error) {
	if r.isSequential {
		return json.Marshal(r.Last())
	} else {
		return json.Marshal(r.steps)
	}
}

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
