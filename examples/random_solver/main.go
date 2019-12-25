package main

import (
	"github.com/sile/kurobako-go"
	"math"
	"math/rand"
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

func (r *randomSolver) Ask(idg *kurobako.TrialIdGenerator) (kurobako.NextTrial, error) {
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

	trial.TrialId = idg.Generate()
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
