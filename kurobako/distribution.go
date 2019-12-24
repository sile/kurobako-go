package kurobako

import (
	"encoding/json"
	"fmt"
)

type Distribution int

const (
	Uniform Distribution = iota
	LogUniform
)

func (r Distribution) String() string {
	switch r {
	case Uniform:
		return "UNIFORM"
	case LogUniform:
		return "LOG_UNIFORM"
	default:
		panic("unknown distribution")
	}
}

func (r Distribution) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Distribution) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "UNIFORM":
		*r = Uniform
	case "LOG_UNIFORM":
		*r = LogUniform
	default:
		return fmt.Errorf("unknown `Distribution`: %s", s)
	}

	return nil
}
