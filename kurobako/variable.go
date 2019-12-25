package kurobako

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"math"
)

type Var struct {
	Name         string       `json:"name"`
	Range        Range        `json:"range"`
	Distribution Distribution `json:"distribution"`
	Constraint   *string      `json:"constraint"`
}

func NewVar(name string) Var {
	return Var{
		Name:         name,
		Range:        ContinuousRange{math.Inf(-1), math.Inf(0)}.ToRange(),
		Distribution: Uniform,
		Constraint:   nil,
	}
}

func (r Var) IsConstraintSatisfied(vars []Var, vals []float64) (bool, error) {
	if r.Constraint == nil {
		return true, nil
	}

	lua_state := lua.NewState()
	defer lua_state.Close()

	for i := 0; i < len(vars) && i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			// This is a conditional variable and has n't been bound a value.
			continue
		}

		if x := vars[i].Range.AsContinuousRange(); x != nil {
			lua_state.SetGlobal(vars[i].Name, lua.LNumber(vals[i]))
		} else if x := vars[i].Range.AsDiscreteRange(); x != nil {
			lua_state.SetGlobal(vars[i].Name, lua.LNumber(int(vals[i])))
		} else if x := vars[i].Range.AsCategoricalRange(); x != nil {
			index := int(vals[i])
			lua_state.SetGlobal(vars[i].Name, lua.LString(x.Choices[index]))
		}
	}

	if err := lua_state.DoString(*r.Constraint); err != nil {
		return false, err
	} else {
		value := lua_state.Get(-1)
		satisfied, ok := value.(lua.LBool)
		if !ok {
			return false, fmt.Errorf("expected a lua bool value, got %v", value)
		}
		return bool(satisfied), nil
	}
}
