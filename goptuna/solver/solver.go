package solver

import (
	"fmt"
	"github.com/c-bata/goptuna"
	"github.com/sile/kurobako-go"
)

type GoptunaSolverFactory struct {
	createStudy func(int64) (*goptuna.Study, error)
}

func NewGoptunaSolverFactory(createStudy func(int64) (*goptuna.Study, error)) GoptunaSolverFactory {
	return GoptunaSolverFactory{createStudy}
}

func (r *GoptunaSolverFactory) Specification() (*kurobako.SolverSpec, error) {
	spec := kurobako.NewSolverSpec("Goptuna")
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
	} else {
		var waitings trialQueue
		var pruned trialQueue
		runnings := map[uint64]int{}
		return &GoptunaSolver{study, problem, waitings, pruned, runnings}, nil
	}
}

type GoptunaSolver struct {
	study    *goptuna.Study
	problem  kurobako.ProblemSpec
	waitings trialQueue
	pruned   trialQueue
	runnings map[uint64]int
}

func (r *GoptunaSolver) Ask(idg *kurobako.TrialIdGenerator) (kurobako.NextTrial, error) {
	var nextTrial kurobako.NextTrial
	var goptunaTrialId int

	nextTrial.Params = []*float64{}
	if item := r.pruned.pop(); item != nil {
		nextTrial.TrialId = item.kurobakoTrialId
		goptunaTrialId = item.goptunaTrialId
		nextTrial.NextStep = 0 // `0` indicates that this trial has been pruned.
	} else if item := r.waitings.pop(); item != nil {
		nextTrial.TrialId = item.kurobakoTrialId
		goptunaTrialId = item.goptunaTrialId

		frozenTrial, err := r.study.Storage.GetTrial(goptunaTrialId)
		if err != nil {
			return nextTrial, err
		}

		currentStep, _ := frozenTrial.GetLatestStep()
		nextTrial.NextStep = uint64(currentStep) + 1
	} else {
		nextTrial.TrialId = idg.Generate()
		newGoptunaTrialId, err := r.study.Storage.CreateNewTrial(r.study.ID)
		if err != nil {
			return nextTrial, err
		}
		goptunaTrialId = newGoptunaTrialId
		nextTrial.NextStep = 1

		for _, p := range r.problem.Params {
			satisfied, err := p.IsConstraintSatisfied(r.problem.Params, nextTrial.Params)
			if err != nil {
				return nextTrial, err
			}

			if satisfied {
				value, err := r.suggest(goptunaTrialId, p)
				if err != nil {
					return nextTrial, err
				}
				nextTrial.Params = append(nextTrial.Params, &value)
			} else {
				nextTrial.Params = append(nextTrial.Params, nil)
			}
		}
	}

	r.runnings[nextTrial.TrialId] = goptunaTrialId
	return nextTrial, nil
}

func (r *GoptunaSolver) Tell(trial kurobako.EvaluatedTrial) error {
	kurobakoTrialId := trial.TrialId
	values := trial.Values
	currentStep := trial.CurrentStep

	goptunaTrialId, ok := r.runnings[kurobakoTrialId]
	if !ok {
		return fmt.Errorf("unknown trial: kurobakoTrialId=%v", kurobakoTrialId)
	}
	delete(r.runnings, kurobakoTrialId)

	if len(values) == 0 {
		return r.study.Storage.SetTrialState(goptunaTrialId, goptuna.TrialStatePruned)
	}

	value := values[0]
	if r.study.Direction() == goptuna.StudyDirectionMaximize {
		value = -value
	}

	if currentStep >= r.problem.Steps.Last() {
		if err := r.study.Storage.SetTrialValue(goptunaTrialId, value); err != nil {
			return err
		}
		return r.study.Storage.SetTrialState(goptunaTrialId, goptuna.TrialStateComplete)
	} else {
		if err := r.study.Storage.SetTrialValue(goptunaTrialId, value); err != nil {
			return err
		}
		if err := r.study.Storage.SetTrialIntermediateValue(goptunaTrialId, int(currentStep), value); err != nil {
			return err
		}

		trial, err := r.study.Storage.GetTrial(goptunaTrialId)
		if err != nil {
			return err
		}

		shouldPrune, err := r.study.Pruner.Prune(r.study, trial)
		if err != nil {
			return err
		}

		if shouldPrune {
			r.pruned.push(trialQueueItem{kurobakoTrialId, goptunaTrialId})
			return r.study.Storage.SetTrialState(goptunaTrialId, goptuna.TrialStatePruned)
		} else {
			r.waitings.push(trialQueueItem{kurobakoTrialId, goptunaTrialId})
			return nil
		}
	}
}

func (r *GoptunaSolver) suggest(goptunaTrialId int, param kurobako.Var) (float64, error) {
	trial, err := r.study.Storage.GetTrial(goptunaTrialId)
	if err != nil {
		return 0.0, err
	}

	if x := param.Range.AsContinuousRange(); x != nil {
		if param.Distribution == kurobako.Uniform {
			distribution := goptuna.UniformDistribution{
				Low:  x.Low,
				High: x.High,
			}
			return r.study.Sampler.Sample(r.study, trial, param.Name, distribution)
		} else {
			distribution := goptuna.LogUniformDistribution{
				Low:  x.Low,
				High: x.High,
			}
			return r.study.Sampler.Sample(r.study, trial, param.Name, distribution)
		}
	} else if x := param.Range.AsDiscreteRange(); x != nil {
		if param.Distribution == kurobako.Uniform {
			distribution := goptuna.IntUniformDistribution{
				Low:  int(x.Low),
				High: int(x.High - 1),
			}
			return r.study.Sampler.Sample(r.study, trial, param.Name, distribution)
		}
	} else if x := param.Range.AsCategoricalRange(); x != nil {
		distribution := goptuna.CategoricalDistribution{Choices: x.Choices}
		return r.study.Sampler.Sample(r.study, trial, param.Name, distribution)
	}

	return 0.0, fmt.Errorf("unsupported parameter: %v", param)
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
	} else {
		item := r.items[0]
		r.items = r.items[1:]
		if len(r.items) == 0 {
			r.items = nil
		}
		return &item
	}
}

type trialQueueItem struct {
	kurobakoTrialId uint64
	goptunaTrialId  int
}
