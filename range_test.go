package kurobako

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRangeMarshalAndUnmarshal(t *testing.T) {
	type Pair struct {
		value Range
		json  string
	}

	data := [...]Pair{
		Pair{ContinuousRange{0.1, 3.5}.ToRange(), "{\"high\":3.5,\"low\":0.1,\"type\":\"CONTINUOUS\"}"},
		Pair{DiscreteRange{-5, 5}.ToRange(), "{\"high\":5,\"low\":-5,\"type\":\"DISCRETE\"}"},
		Pair{CategoricalRange{[]string{"foo", "bar", "baz"}}.ToRange(), "{\"choices\":[\"foo\",\"bar\",\"baz\"],\"type\":\"CATEGORICAL\"}"},
	}

	for _, p := range data {
		bytes, err := json.Marshal(p.value)
		if err != nil {
			t.Fatal(err)
		}

		text := string(bytes)
		if text != p.json {
			t.Fatalf("unexpected JSON: %v", text)
		}

		var x Range
		err = json.Unmarshal([]byte(p.json), &x)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(x, p.value) {
			t.Fatalf("unexpected Distribution: %v", x)
		}
	}
}
