package kurobako

import (
	"encoding/json"
)

type Range interface {
	Low() float64
 	High() float64
}

type ContinuousRange struct {
	low float64 `json:"low"`
	high float64 `json:"high"`
}

func (r *ContinuousRange) Low() float64 {
	return r.low
}
