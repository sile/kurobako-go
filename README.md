kurobako-go
===========

![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)
[![GoDoc](https://godoc.org/github.com/sile/kurobako-go?status.svg)](https://godoc.org/github.com/sile/kurobako-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sile/kurobako-go)](https://goreportcard.com/report/github.com/sile/kurobako-go)
[![Actions Status](https://github.com/sile/kurobako-go/workflows/CI/badge.svg)](https://github.com/sile/kurobako-go/actions)

A Golang library to help implement [kurobako]'s solvers and problems.

[kurobako]: https://github.com/sile/kurobako


Usage Examples
--------------

### 1. Define a solver based on random search

```go
// file: random-solver.go
package main

import (
	"math"
	"math/rand"

	"github.com/sile/kurobako-go"
)

type randomSolverFactory struct{}

func (r *randomSolverFactory) Specification() (*kurobako.SolverSpec, error) {
	spec := kurobako.NewSolverSpec("Random Search")
	return &spec, nil
}

func (r *randomSolverFactory) CreateSolver(seed int64, problem kurobako.ProblemSpec) (kurobako.Solver, error) {
	rng := rand.New(rand.NewSource(seed))
	return &randomSolver{rng, problem}, nil
}

type randomSolver struct {
	rng     *rand.Rand
	problem kurobako.ProblemSpec
}

func (r *randomSolver) sampleUniform(low float64, high float64) float64 {
	return r.rng.Float64()*(high-low) + low
}

func (r *randomSolver) sampleLogUniform(low float64, high float64) float64 {
	return math.Exp(r.sampleUniform(math.Log(low), math.Log(high)))
}

func (r *randomSolver) Ask(idg *kurobako.TrialIDGenerator) (kurobako.NextTrial, error) {
	var trial kurobako.NextTrial

	for _, p := range r.problem.Params {
		if p.Distribution == kurobako.Uniform {
			value := r.sampleUniform(p.Range.Low(), p.Range.High())
			trial.Params = append(trial.Params, &value)
		} else {
			value := r.sampleLogUniform(p.Range.Low(), p.Range.High())
			trial.Params = append(trial.Params, &value)
		}
	}

	trial.TrialID = idg.Generate()
	trial.NextStep = r.problem.Steps.Last()
	return trial, nil
}

func (r *randomSolver) Tell(trial kurobako.EvaluatedTrial) error {
	return nil
}

func main() {
	runner := kurobako.NewSolverRunner(&randomSolverFactory{})
	if err := runner.Run(); err != nil {
		panic(err)
	}
}
```

### 2. Define a solver based on [Goptuna]

[Goptuna]: https://github.com/c-bata/goptuna

### 3. Define a problem that represents a quadratic function `x**2 + y`

### 4. Run a benchmark that uses the above solver and problem
