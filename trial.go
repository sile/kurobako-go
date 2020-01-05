package kurobako

import (
	"encoding/json"
)

// NextTrial contains information about a trial to be evaluated.
type NextTrial struct {
	// TrialID is the identifier of the trial.
	TrialID  uint64     `json:"id"`

	// Params are the parameters to be evaluated.
	Params   []*float64 `json:"params"`

	// NextStep is the next evaluable step.
	//
	// The evaluator must go through the next evaluation process beyond this step.
	// Note that if the value is 0, it means the trial was pruned by solver.
	NextStep uint64     `json:"next_step"`
}

// MarshalJSON encodes a NextTrial object to JSON bytes.
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

// EvaluatedTrial contains information about an evaluated trial.
type EvaluatedTrial struct {
	// TrialID is the identifier of the trial.
	TrialID     uint64    `json:"id"`

	// Values is the evaluation result of the trial.
	//
	// Note that if this is an empty slice, it means the trial contained an unevalable parameter set.
	Values      []float64 `json:"values"`

	// CurrentStep is the current step of the evaluation process.
	CurrentStep uint64    `json:"current_step"`
}

// TrialIDGenerator generates identifiers of trials.
type TrialIDGenerator struct {
	NextID uint64
}

// Generate creates a new trial identifier.
func (r *TrialIDGenerator) Generate() uint64 {
	id := r.NextID
	r.NextID++
	return id
}
