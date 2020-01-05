package kurobako

import (
	"encoding/json"
	"fmt"
)

// Distribution of the values of a parameter.
type Distribution int

const (
	// Uniform indicates the values of the parameter are uniformally distributed.
	Uniform Distribution = iota

	// LogUniform indicates the values of the parameter are log-uniformally distributed.
	LogUniform
)

// String returns the string representation of a Distribution value.
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

// MarshalJSON encodes a Distribution value to JSON bytes.
func (r Distribution) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON decodes a Distribution value from JSON bytes.
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
