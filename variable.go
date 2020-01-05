package kurobako

import (
	"fmt"
	"math"

	lua "github.com/yuin/gopher-lua"
)

// Var is a definition of a variable.
type Var struct {
	// Name is the name of the variable.
	Name         string       `json:"name"`

	// Range is the value range of the variable.
	Range        Range        `json:"range"`

	// Distribution is the value distribution of the variable.
	Distribution Distribution `json:"distribution"`

	// Constraint is the constraint of the variable.
	//
	// A constraint is represented by a Lua script.
	// If the script returns true, it means the constraint is satisfied.
	//
 	// If the constraint isn't satisfied, the variable won't be considered during the evaluation process.
	Constraint   *string      `json:"constraint"`
}

// NewVar creates a new Var instance.
func NewVar(name string) Var {
	return Var{
		Name:         name,
		Range:        ContinuousRange{math.Inf(-1), math.Inf(0)}.ToRange(),
		Distribution: Uniform,
		Constraint:   nil,
	}
}

// IsConstraintSatisfied checks whether the constraint of the variable is satisfied under the given bound (i.e., already evaluated) variables.
func (r Var) IsConstraintSatisfied(vars []Var, vals []*float64) (bool, error) {
	if r.Constraint == nil {
		return true, nil
	}

	luaState := lua.NewState()
	defer luaState.Close()

	for i := 0; i < len(vars) && i < len(vals); i++ {
		if vals[i] == nil {
			// This is a conditional variable and has n't been bound a value.
			continue
		}

		if x := vars[i].Range.AsContinuousRange(); x != nil {
			luaState.SetGlobal(vars[i].Name, lua.LNumber(*vals[i]))
		} else if x := vars[i].Range.AsDiscreteRange(); x != nil {
			luaState.SetGlobal(vars[i].Name, lua.LNumber(int(*vals[i])))
		} else if x := vars[i].Range.AsCategoricalRange(); x != nil {
			index := int(*vals[i])
			luaState.SetGlobal(vars[i].Name, lua.LString(x.Choices[index]))
		}
	}

	if err := luaState.DoString(*r.Constraint); err != nil {
		return false, err
	}

	value := luaState.Get(-1)
	satisfied, ok := value.(lua.LBool)
	if !ok {
		return false, fmt.Errorf("expected a lua bool value, got %v", value)
	}
	return bool(satisfied), nil
}
