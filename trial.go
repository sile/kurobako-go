package kurobako

import (
	"encoding/json"
)

type NextTrial struct {
	TrialID  uint64     `json:"id"`
	Params   []*float64 `json:"params"`
	NextStep uint64     `json:"next_step"`
}

func (r NextTrial) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"id":        r.TrialID,
		"params":    r.Params,
		"next_step": r.NextStep,
	}
	if r.NextStep == 0 {
		delete(m, "next_step")
	}
	return json.Marshal(m)
}

type EvaluatedTrial struct {
	TrialID     uint64    `json:"id"`
	Values      []float64 `json:"values"`
	CurrentStep uint64    `json:"current_step"`
}

type TrialIDGenerator struct {
	NextID uint64
}

func (r *TrialIDGenerator) Generate() uint64 {
	id := r.NextID
	r.NextID++
	return id
}
