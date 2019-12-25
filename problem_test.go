package kurobako

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProglemSpecMarshlAndUnmarshal(t *testing.T) {
	spec := NewProblemSpec("Quadratic Function")

	x := NewVar("x")
	x.Range = ContinuousRange{-10.0, 10.0}.ToRange()

	y := NewVar("y")
	y.Range = DiscreteRange{-3, 3}.ToRange()

	spec.Params = []Var{x, y}

	spec.Values = []Var{NewVar("x**2 + y")}

	bytes, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}

	var spec2 ProblemSpec
	err = json.Unmarshal(bytes, &spec2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec, spec2) {
		t.Fatalf("unexpected `ProblemSpec`: %v (JSON=%s)", spec2, string(bytes))
	}

	text := "{\"name\":\"Quadratic Function\",\"attrs\":{},\"params_domain\":[{\"name\":\"x\",\"range\":{\"type\":\"CONTINUOUS\",\"low\":-10.0,\"high\":10.0},\"distribution\":\"UNIFORM\",\"constraint\":null},{\"name\":\"y\",\"range\":{\"type\":\"DISCRETE\",\"low\":-3,\"high\":3},\"distribution\":\"UNIFORM\",\"constraint\":null}],\"values_domain\":[{\"name\":\"x**2 + y\",\"range\":{\"type\":\"CONTINUOUS\"},\"distribution\":\"UNIFORM\",\"constraint\":null}],\"steps\":1}"

	var spec3 ProblemSpec
	err = json.Unmarshal([]byte(text), &spec3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec, spec3) {
		t.Fatalf("unexpected `ProblemSpec`: %v (JSON=%s)", spec3, text)
	}
}
