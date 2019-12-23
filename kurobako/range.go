package kurobako

import (
	"encoding/json"
	"fmt"
)

type ContinuousRange struct {
	Low  float64
	High float64
}

func (r ContinuousRange) ToRange() Range {
	return Range{r}
}

type DiscreteRange struct {
	Low  int64
	High int64
}

func (r DiscreteRange) ToRange() Range {
	return Range{r}
}

type CategoricalRange struct {
	Choices []string
}

func (r CategoricalRange) ToRange() Range {
	return Range{r}
}

type Range struct {
	inner interface{}
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
		return json.Marshal(map[string]interface{}{
			"type": "CONTINUOUS",
			"low":  x.Low,
			"high": x.High,
		})
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
		low, ok := m["low"].(float64)
		if !ok {
			return fmt.Errorf("\"low\" should be a float: %v", m["low"])
		}

		high, ok := m["high"].(float64)
		if !ok {
			return fmt.Errorf("\"high\" should be a float: %v", m["high"])
		}

		*r = ContinuousRange{low, high}.ToRange()

	case "DISCRETE":
		low, ok := m["low"].(int64)
		if !ok {
			return fmt.Errorf("\"low\" should be a int: %v", m["low"])
		}

		high, ok := m["high"].(int64)
		if !ok {
			return fmt.Errorf("\"high\" should be a int: %v", m["high"])
		}

		*r = DiscreteRange{low, high}.ToRange()

	case "CATEGORICAL":
		choices, ok := m["choices"].([]string)
		if !ok {
			return fmt.Errorf("\"choices\" should be a list of string: %v", m["choices"])
		}

		*r = CategoricalRange{choices}.ToRange()

	default:
		return fmt.Errorf("unknown or missing \"type\" field: %v", m["type"])
	}

	return nil
}
