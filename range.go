package kurobako

import (
	"encoding/json"
	"fmt"
	"math"
)

func isFinite(v float64) bool {
	return !(math.IsInf(v, 0) || math.IsNaN(v))
}

// ContinuousRange represents a numerical continuous range.
type ContinuousRange struct {
	// Low is the lower bound of the range (inclusive).
	Low float64 `json:"low"`

	// High is the upper bound of the range (exclusive).
	High float64 `json:"high"`
}

// ToRange creates a Range object that contains the receiver object.
func (r ContinuousRange) ToRange() Range {
	return Range{r}
}

// MarshalJSON encodes a ContinuousRange object to JSON bytes.
func (r ContinuousRange) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type": "CONTINUOUS",
		"low":  r.Low,
		"high": r.High,
	}

	if !isFinite(r.Low) {
		delete(m, "low")
	}
	if !isFinite(r.High) {
		delete(m, "high")
	}

	return json.Marshal(m)
}

// UnmarshalJSON decodes a ContinuousRange object from JSON bytes.
func (r *ContinuousRange) UnmarshalJSON(data []byte) error {
	var m struct {
		Low  *float64 `json:"low"`
		High *float64 `json:"high"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if m.Low == nil {
		r.Low = math.Inf(-1)
	} else {
		r.Low = *m.Low
	}

	if m.High == nil {
		r.High = math.Inf(0)
	} else {
		r.High = *m.High
	}

	return nil
}

// DiscreteRange represents a numerical discrete range.
type DiscreteRange struct {
	// Low is the lower bound of the range (inclusive).
	Low int64 `json:"low"`

	// High is the upper bound of the range (exclusive).
	High int64 `json:"high"`
}

// ToRange create a Range object that contains the receiver object.
func (r DiscreteRange) ToRange() Range {
	return Range{r}
}

// CategoricalRange represents a categorical range (choices).
type CategoricalRange struct {
	// Choices is the possible values in the range.
	Choices []string `json:"choices"`
}

// ToRange creates a Range object that contains the receiver object.
func (r CategoricalRange) ToRange() Range {
	return Range{r}
}

// Range represents the range of a parameter.
type Range struct {
	inner interface{}
}

// Low is the lower bound of the range (inclusive).
func (r *Range) Low() float64 {
	switch x := (r.inner).(type) {
	case ContinuousRange:
		return x.Low
	case DiscreteRange:
		return float64(x.Low)
	case CategoricalRange:
		return 0.0
	default:
		panic("unreachable")
	}
}

// High is the upper bound of the range (exclusive).
func (r *Range) High() float64 {
	switch x := (r.inner).(type) {
	case ContinuousRange:
		return x.High
	case DiscreteRange:
		return float64(x.High)
	case CategoricalRange:
		return float64(len(x.Choices))
	default:
		panic("unreachable")
	}
}

// AsContinuousRange tries to return the inner object of the range as a ContinuousRange object.
func (r *Range) AsContinuousRange() *ContinuousRange {
	inner, ok := (r.inner).(ContinuousRange)
	if ok {
		return &inner
	}
	return nil
}

// AsDiscreteRange tries to return the inner object of the range as a DiscreteRange object.
func (r *Range) AsDiscreteRange() *DiscreteRange {
	inner, ok := (r.inner).(DiscreteRange)
	if ok {
		return &inner
	}
	return nil
}

// AsCategoricalRange tries to return the inner object of the range as a CategoricalRange object.
func (r *Range) AsCategoricalRange() *CategoricalRange {
	inner, ok := (r.inner).(CategoricalRange)
	if ok {
		return &inner
	}
	return nil
}

// MarshalJSON encodes a range object to JSON bytes.
func (r Range) MarshalJSON() ([]byte, error) {
	if x := r.AsContinuousRange(); x != nil {
		return json.Marshal(x)
	} else if x := r.AsDiscreteRange(); x != nil {
		return json.Marshal(map[string]interface{}{
			"type": "DISCRETE",
			"low":  x.Low,
			"high": x.High,
		})
	} else if x := r.AsCategoricalRange(); x != nil {
		return json.Marshal(map[string]interface{}{
			"type":    "CATEGORICAL",
			"choices": x.Choices,
		})
	} else {
		panic("unreachable")
	}
}

// UnmarshalJSON decodes a Range object from JSON bytes.
func (r *Range) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	switch m["type"] {
	case "CONTINUOUS":
		var x ContinuousRange
		if err := json.Unmarshal(data, &x); err != nil {
			return err
		}
		*r = x.ToRange()

	case "DISCRETE":
		var x DiscreteRange
		if err := json.Unmarshal(data, &x); err != nil {
			return err
		}
		*r = x.ToRange()

	case "CATEGORICAL":
		var x CategoricalRange
		if err := json.Unmarshal(data, &x); err != nil {
			return err
		}
		*r = x.ToRange()

	default:
		return fmt.Errorf("unknown or missing \"type\" field: %v", m["type"])
	}

	return nil
}
