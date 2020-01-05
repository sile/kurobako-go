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

```go
// file: goptuna-solver.go
package main

import (
	"github.com/c-bata/goptuna"
	"github.com/c-bata/goptuna/tpe"
	"github.com/sile/kurobako-go"
	"github.com/sile/kurobako-go/goptuna/solver"
)

func createStudy(seed int64) (*goptuna.Study, error) {
	sampler := tpe.NewSampler(tpe.SamplerOptionSeed(seed))
	return goptuna.CreateStudy("example-study", goptuna.StudyOptionSampler(sampler))
}

func main() {
	factory := solver.NewGoptunaSolverFactory(createStudy)
	runner := kurobako.NewSolverRunner(&factory)
	if err := runner.Run(); err != nil {
		panic(err)
	}
}
```

### 3. Define a problem that represents a quadratic function `x**2 + y`

```go
// file: quadratic-problem.go
package main

import (
	"github.com/sile/kurobako-go"
)

type quadraticProblemFactory struct {
}

func (r *quadraticProblemFactory) Specification() (*kurobako.ProblemSpec, error) {
	spec := kurobako.NewProblemSpec("Quadratic Function")

	x := kurobako.NewVar("x")
	x.Range = kurobako.ContinuousRange{-10.0, 10.0}.ToRange()

	y := kurobako.NewVar("y")
	y.Range = kurobako.DiscreteRange{-3, 3}.ToRange()

	spec.Params = []kurobako.Var{x, y}

	spec.Values = []kurobako.Var{kurobako.NewVar("x**2 + y")}

	return &spec, nil
}

func (r *quadraticProblemFactory) CreateProblem(seed int64) (kurobako.Problem, error) {
	return &quadraticProblem{}, nil
}

type quadraticProblem struct {
}

func (r *quadraticProblem) CreateEvaluator(params []float64) (kurobako.Evaluator, error) {
	x := params[0]
	y := params[1]
	return &quadraticEvaluator{x, y}, nil
}

type quadraticEvaluator struct {
	x float64
	y float64
}

func (r *quadraticEvaluator) Evaluate(nextStep uint64) (uint64, []float64, error) {
	values := []float64{r.x*r.x + r.y}
	return 1, values, nil
}

func main() {
	runner := kurobako.NewProblemRunner(&quadraticProblemFactory{})
	if err := runner.Run(); err != nil {
		panic(err)
	}
}
```

### 4. Run a benchmark that uses the above solver and problem

```console
// Define solver and problem.
$ SOLVER1=$(kurobako solver command go run random-solver.go)
$ SOLVER2=$(kurobako solver command go run goptuna-solver.go)
$ PROBLEM=$(kurobako problem command go run quadratic-problem.go)

// Execute benchmark.
$ kurobako studies --solvers $SOLVER1 $SOLVER2 --problems $PROBLEM | kurobako run > result.json

// Generate Markdown format report and visualization image.
$ cat result.json | kurobako report

$ cat result.json | kurobako plot curve
```
