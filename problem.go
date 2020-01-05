package kurobako

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrorUnevalableParams is an error that is used when an evaluator encounters an infeasible parameter set.
var ErrorUnevalableParams = errors.New("unevalable params")

// ProblemSpec is the specification of a black-box optimization problem.
type ProblemSpec struct {
	// Name is the name of the problem.
	Name string `json:"name"`

	// Attrs is the attributes of the problem.
	Attrs map[string]string `json:"attrs"`

	// Params is the definition of the parameters domain of the problem.
	Params []Var `json:"params_domain"`

	// Values is the definition of the values domain of the problem.
	Values []Var `json:"values_domain"`

	// Steps is the sequence of the evaluation steps of the problem.
	Steps Steps `json:"steps"`
}

// NewProblemSpec creates a new ProblemSpec instance.
func NewProblemSpec(name string) ProblemSpec {
	steps, _ := NewSteps([]uint64{1})
	return ProblemSpec{
		Name:  name,
		Attrs: map[string]string{},
		Steps: *steps,
	}
}

// Evaluator allows to execute an evaluation process.
type Evaluator interface {
	// evaluate executes an evaluation process, at least, until the given step.
	Evaluate(nextStep uint64) (currentStep uint64, values []float64, err error)
}

// Problem allows to create a new evaluator instance.
type Problem interface {
	// CreateEvaluator creates a new evaluator to evaluate the given parameter set.
	CreateEvaluator(params []float64) (Evaluator, error)
}

// ProblemFactory allows to create a new problem instance.
type ProblemFactory interface {
	// Specification returns the specification of the problem.
	Specification() (*ProblemSpec, error)

	// CreateProblem creates a new problem instance with the given random seed.
	CreateProblem(seed int64) (Problem, error)
}

// ProblemRunner runs a black-box optimization problem.
type ProblemRunner struct {
	factory    ProblemFactory
	problems   map[uint64]Problem
	evaluators map[uint64]Evaluator
}

// NewProblemRunner creates a new ProblemRunner that runs the given problem.
func NewProblemRunner(factory ProblemFactory) *ProblemRunner {
	return &ProblemRunner{factory, nil, nil}
}

// Run runs the problem.
func (r *ProblemRunner) Run() error {
	r.problems = map[uint64]Problem{}
	r.evaluators = map[uint64]Evaluator{}

	if err := r.castProblemSpec(); err != nil {
		return err
	}

	for {
		doContinue, err := r.runOnce()
		if err != nil {
			return err
		}

		if !doContinue {
			break
		}
	}

	return nil
}

func (r *ProblemRunner) runOnce() (bool, error) {
	line, err := readLine()
	if err == io.EOF {
		return false, nil
	} else if err != nil {
		return false, err
	}

	var message map[string]interface{}
	if err := json.Unmarshal(line, &message); err != nil {
		return false, err
	}

	switch message["type"] {
	case "CREATE_PROBLEM_CAST":
		err := r.handleCreateProblemCast(line)
		return true, err
	case "DROP_PROBLEM_CAST":
		err := r.handleDropProblemCast(line)
		return true, err
	case "CREATE_EVALUATOR_CALL":
		err := r.handleCreateEvaluatorCall(line)
		return true, err
	case "DROP_EVALUATOR_CAST":
		err := r.handleDropEvaluatorCast(line)
		return true, err
	case "EVALUATE_CALL":
		err := r.handleEvaluateCall(line)
		return true, err
	default:
		return false, fmt.Errorf("unknown message type: %v", message["type"])
	}
}

func (r *ProblemRunner) handleEvaluateCall(input []byte) error {
	var message struct {
		EvaluatorID uint64 `json:"evaluator_id"`
		NextStep    uint64 `json:"next_step"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	evaluator := r.evaluators[message.EvaluatorID]
	currentStep, values, err := evaluator.Evaluate(message.NextStep)
	if err != nil {
		return err
	}

	reply := map[string]interface{}{
		"type":         "EVALUATE_REPLY",
		"current_step": currentStep,
		"values":       values,
	}
	return r.sendMessage(reply)
}

func (r *ProblemRunner) handleDropEvaluatorCast(input []byte) error {
	var message struct {
		EvaluatorID uint64 `json:"evaluator_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	delete(r.evaluators, message.EvaluatorID)
	return nil
}

func (r *ProblemRunner) handleCreateEvaluatorCall(input []byte) error {
	var message struct {
		ProblemID   uint64    `json:"problem_id"`
		EvaluatorID uint64    `json:"evaluator_id"`
		Params      []float64 `json:"params"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	problem := r.problems[message.ProblemID]
	evaluator, err := problem.CreateEvaluator(message.Params)
	if err == ErrorUnevalableParams {
		return r.sendMessage(map[string]interface{}{"type": "ERROR_REPLY", "kind": "UNEVALABLE_PARAMS"})
	} else if err != nil {
		return err
	}

	r.evaluators[message.EvaluatorID] = evaluator
	return r.sendMessage(map[string]interface{}{"type": "CREATE_EVALUATOR_REPLY"})
}

func (r *ProblemRunner) handleDropProblemCast(input []byte) error {
	var message struct {
		ProblemID uint64 `json:"problem_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	delete(r.problems, message.ProblemID)
	return nil
}

func (r *ProblemRunner) handleCreateProblemCast(input []byte) error {
	var message struct {
		ProblemID  uint64 `json:"problem_id"`
		RandomSeed uint64 `json:"random_seed"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	problem, err := r.factory.CreateProblem(int64(message.RandomSeed))
	if err != nil {
		return err
	}

	r.problems[message.ProblemID] = problem
	return nil
}

func (r *ProblemRunner) castProblemSpec() error {
	spec, err := r.factory.Specification()
	if err != nil {
		return err
	}

	return r.sendMessage(map[string]interface{}{"type": "PROBLEM_SPEC_CAST", "spec": spec})
}

func (r *ProblemRunner) sendMessage(message map[string]interface{}) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(bytes))
	return nil
}

var stdin = bufio.NewReaderSize(os.Stdin, 1024*1024)

func readLine() ([]byte, error) {
	line, prefix, err := stdin.ReadLine()

	if err != nil {
		return nil, err
	}

	if prefix {
		return nil, fmt.Errorf("Too long input")
	}

	return line, nil
}
