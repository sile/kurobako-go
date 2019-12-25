package kurobako

import (
	"encoding/json"
	"testing"
)

func TestDistributionMarshlAndUnmarshal(t *testing.T) {
	type Pair struct {
		value Distribution
		json  string
	}

	data := [...]Pair{
		Pair{Uniform, "\"UNIFORM\""},
		Pair{LogUniform, "\"LOG_UNIFORM\""},
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

		var d Distribution
		err = json.Unmarshal([]byte(p.json), &d)
		if err != nil {
			t.Fatal(err)
		}
		if d != p.value {
			t.Fatalf("unexpected Distribution: %v", d)
		}
	}
}
