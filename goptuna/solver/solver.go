// This package provides a solver base on Goptuna.
//
// Please see https://github.com/c-bata/goptuna for Goptuna.
package solver

import (
	"fmt"

	"github.com/c-bata/goptuna"
	"github.com/sile/kurobako-go"
)

// GoptunaSolverFactory is a SolverFactory for Goptuna.
type GoptunaSolverFactory struct {
	name        string
	createStudy func(int64) (*goptuna.Study, error)
}

// NewGoptunaSolverFactory creates a new GoptunaSolverFactory instance.
//
// The createStudy argument is a function that takes a random seed and returns a study object that
// is used to solve black-box optimization problems.
func NewGoptunaSolverFactory(name string, createStudy func(int64) (*goptuna.Study, error)) GoptunaSolverFactory {
	if name == "" {
		name = "Goptuna"
	}
	return GoptunaSolverFactory{
		name:        name,
		createStudy: createStudy,
	}
}

func (r *GoptunaSolverFactory) Specification() (*kurobako.SolverSpec, error) {
	spec := kurobako.NewSolverSpec(r.name)
	spec.Attrs["github"] = "https://github.com/c-bata/goptuna"
	spec.Capabilities = kurobako.UniformContinuous |
		kurobako.LogUniformContinuous |
		kurobako.UniformDiscrete |
		kurobako.Categorical |
		kurobako.Conditional |
		kurobako.Concurrent
	return &spec, nil
}

func (r *GoptunaSolverFactory) CreateSolver(seed int64, problem kurobako.ProblemSpec) (kurobako.Solver, error) {
	study, err := r.createStudy(seed)
	if err != nil {
		return nil, err
	}

	var waitings trialQueue
	var pruned trialQueue
	runnings := map[uint64]int{}
	return &GoptunaSolver{study, problem, waitings, pruned, runnings}, nil
}

// GoptunaSolver is a Solver implementation based on Goptuna.
type GoptunaSolver struct {
	study    *goptuna.Study
	problem  kurobako.ProblemSpec
	waitings trialQueue
	pruned   trialQueue
	runnings map[uint64]int
}

func (r *GoptunaSolver) Ask(idg *kurobako.TrialIDGenerator) (kurobako.NextTrial, error) {
	var nextTrial kurobako.NextTrial
	var goptunaTrialID int

	nextTrial.Params = []*float64{}
	if item := r.pruned.pop(); item != nil {
		nextTrial.TrialID = item.kurobakoTrialID
		goptunaTrialID = item.goptunaTrialID
		nextTrial.NextStep = 0 // `0` indicates that this trial has been pruned.
	} else if item := r.waitings.pop(); item != nil {
		nextTrial.TrialID = item.kurobakoTrialID
		goptunaTrialID = item.goptunaTrialID

		frozenTrial, err := r.study.Storage.GetTrial(goptunaTrialID)
		if err != nil {
			return nextTrial, err
		}

		currentStep, _ := frozenTrial.GetLatestStep()
		nextTrial.NextStep = uint64(currentStep) + 1
	} else {
		nextTrial.TrialID = idg.Generate()
		newGoptunaTrialID, err := r.study.Storage.CreateNewTrial(r.study.ID)
		if err != nil {
			return nextTrial, err
		}
		goptunaTrialID = newGoptunaTrialID
		nextTrial.NextStep = 1

		relativeValues, err := r.callRelativeSampler(goptunaTrialID, r.problem.Params)
		if err != nil {
			return nextTrial, err
		}

		for _, p := range r.problem.Params {
			satisfied, err := p.IsConstraintSatisfied(r.problem.Params, nextTrial.Params)
			if err != nil {
				return nextTrial, err
			}

			if satisfied {
				v, ok := relativeValues[p.Name]
				if ok {
					nextTrial.Params = append(nextTrial.Params, &v)
				} else {
					value, err := r.suggest(goptunaTrialID, p)
					if err != nil {
						return nextTrial, err
					}
					nextTrial.Params = append(nextTrial.Params, &value)
				}
			} else {
				nextTrial.Params = append(nextTrial.Params, nil)
			}
		}
	}

	r.runnings[nextTrial.TrialID] = goptunaTrialID
	return nextTrial, nil
}

func (r *GoptunaSolver) Tell(trial kurobako.EvaluatedTrial) error {
	kurobakoTrialID := trial.TrialID
	values := trial.Values
	currentStep := trial.CurrentStep

	goptunaTrialID, ok := r.runnings[kurobakoTrialID]
	if !ok {
		return fmt.Errorf("unknown trial: kurobakoTrialID=%v", kurobakoTrialID)
	}
	delete(r.runnings, kurobakoTrialID)

	if len(values) == 0 {
		return r.study.Storage.SetTrialState(goptunaTrialID, goptuna.TrialStatePruned)
	}

	value := values[0]
	if r.study.Direction() == goptuna.StudyDirectionMaximize {
		value = -value
	}

	if currentStep >= r.problem.Steps.Last() {
		if err := r.study.Storage.SetTrialValue(goptunaTrialID, value); err != nil {
			return err
		}
		return r.study.Storage.SetTrialState(goptunaTrialID, goptuna.TrialStateComplete)
	}

	if err := r.study.Storage.SetTrialValue(goptunaTrialID, value); err != nil {
		return err
	}
	if err := r.study.Storage.SetTrialIntermediateValue(goptunaTrialID, int(currentStep), value); err != nil {
		return err
	}

	goptunaTrial, err := r.study.Storage.GetTrial(goptunaTrialID)
	if err != nil {
		return err
	}

	shouldPrune, err := r.study.Pruner.Prune(r.study, goptunaTrial)
	if err != nil {
		return err
	}

	if shouldPrune {
		r.pruned.push(trialQueueItem{kurobakoTrialID, goptunaTrialID})
		return r.study.Storage.SetTrialState(goptunaTrialID, goptuna.TrialStatePruned)
	}

	r.waitings.push(trialQueueItem{kurobakoTrialID, goptunaTrialID})
	return nil
}

func (r *GoptunaSolver) callRelativeSampler(goptunaTrialID int, params []kurobako.Var) (map[string]float64, error) {
	if r.study.RelativeSampler == nil {
		return nil, nil
	}

	trial, err := r.study.Storage.GetTrial(goptunaTrialID)
	if err != nil {
		return nil, err
	}

	searchSpace := make(map[string]interface{}, len(params))
	for _, p := range params {
		distribution, err := toGoptunaDistribution(p)
		if err != nil {
			return nil, err
		}
		searchSpace[p.Name] = distribution
	}

	values, err := r.study.RelativeSampler.SampleRelative(r.study, trial, searchSpace)
	if err != nil {
		return nil, err
	}

	for _, p := range params {
		v, ok := values[p.Name]
		if !ok {
			continue
		}

		distribution := searchSpace[p.Name]
		err = r.study.Storage.SetTrialParam(goptunaTrialID, p.Name, v, distribution)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (r *GoptunaSolver) suggest(goptunaTrialID int, param kurobako.Var) (float64, error) {
	trial, err := r.study.Storage.GetTrial(goptunaTrialID)
	if err != nil {
		return 0.0, err
	}

	distribution, err := toGoptunaDistribution(param)
	if err != nil {
		return 0.0, err
	}

	value, err := r.study.Sampler.Sample(r.study, trial, param.Name, distribution)
	if err != nil {
		return 0.0, err
	}

	err = r.study.Storage.SetTrialParam(goptunaTrialID, param.Name, value, distribution)
	if err != nil {
		return 0.0, err
	}

	return value, nil
}

func toGoptunaDistribution(param kurobako.Var) (interface{}, error) {
	if x := param.Range.AsContinuousRange(); x != nil {
		if param.Distribution == kurobako.Uniform {
			return goptuna.UniformDistribution{
				Low:  x.Low,
				High: x.High,
			}, nil
		} else {
			return goptuna.LogUniformDistribution{
				Low:  x.Low,
				High: x.High,
			}, nil
		}
	} else if x := param.Range.AsDiscreteRange(); x != nil {
		if param.Distribution == kurobako.Uniform {
			return goptuna.IntUniformDistribution{
				Low:  int(x.Low),
				High: int(x.High - 1),
			}, nil
		}
	} else if x := param.Range.AsCategoricalRange(); x != nil {
		return goptuna.CategoricalDistribution{Choices: x.Choices}, nil
	}
	return nil, fmt.Errorf("unsupported parameter: %v", param)
}

type trialQueue struct {
	items []trialQueueItem
}

func (r *trialQueue) push(item trialQueueItem) {
	r.items = append(r.items, item)
}

func (r *trialQueue) pop() *trialQueueItem {
	if len(r.items) == 0 {
		return nil
	}

	item := r.items[0]
	r.items = r.items[1:]
	if len(r.items) == 0 {
		r.items = nil
	}
	return &item
}

type trialQueueItem struct {
	kurobakoTrialID uint64
	goptunaTrialID  int
}
