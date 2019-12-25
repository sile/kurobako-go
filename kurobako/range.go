package kurobako

import (
	"encoding/json"
	"fmt"
	"math"
)

func isFinite(v float64) bool {
	return !(math.IsInf(v, 0) || math.IsNaN(v))
}

type ContinuousRange struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

func (r ContinuousRange) ToRange() Range {
	return Range{r}
}

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

func (r *ContinuousRange) UnmarshalJSON(data []byte) error {
	var m struct {
		Low *float64 `json:"low"`
		High *float64 `json:"high"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if m.Low ==nil {
		r.Low  = math.Inf(-1)
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

type DiscreteRange struct {
	Low  int64 `json:"low"`
	High int64 `json:"high"`
}

func (r DiscreteRange) ToRange() Range {
	return Range{r}
}

type CategoricalRange struct {
	Choices []string `json:"choices"`
}

func (r CategoricalRange) ToRange() Range {
	return Range{r}
}

type Range struct {
	inner interface{}
}

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

func (r *Range) AsContinuousRange() *ContinuousRange {
	inner, ok := (r.inner).(ContinuousRange)
	if ok {
		return &inner
	} else {
		return nil
	}
}

func (r *Range) AsDiscreteRange() *DiscreteRange {
	inner, ok := (r.inner).(DiscreteRange)
	if ok {
		return &inner
	} else {
		return nil
	}
}

func (r *Range) AsCategoricalRange() *CategoricalRange {
	inner, ok := (r.inner).(CategoricalRange)
	if ok {
		return &inner
	} else {
		return nil
	}
}

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
