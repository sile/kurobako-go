package kurobako

import (
	"encoding/json"
)

type Range interface {
}

type ContinuousRange struct {
	low float64
	high float64
}

func NewContinuousRange(low float64, high float64) *ContinuousRange {
	if low >= high {
		return nil
	}

	r := ContinuousRange{low, high}
	return &r
}

func (r *ContinuousRange) Low() float64 {
	return r.low
}

func (r *ContinuousRange) High() float64 {
	return r.high
}

type DiscreteRange struct {
	low int64
	high int64
}

func NewDiscreteRange(low int64, high int64) *DiscreteRange {
	if low >= high {
		return nil
	}

	r := DiscreteRange{low, high}
	return &r
}

func (r *DiscreteRange) Low() int64 {
	return r.low
}

func (r *DiscreteRange) High() int64 {
	return r.high
}
