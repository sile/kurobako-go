package kurobako

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var UnevalableParamsError = errors.New("unevalable params")

type ProblemSpec struct {
	Name   string            `json:"name"`
	Attrs  map[string]string `json:"attrs"`
	Params []Var             `json:"params_domain"`
	Values []Var             `json:"values_domain"`
	Steps  Steps             `json:"steps"`
}

func NewProblemSpec(name string) ProblemSpec {
	steps, _ := NewSteps([]uint64{1})
	return ProblemSpec{
		Name:  name,
		Attrs: map[string]string{},
		Steps: *steps,
	}
}

type Evaluator interface {
	Evaluate(nextStep uint64) (currentStep uint64, values []float64, err error)
}

type Problem interface {
	CreateEvaluator(params []float64) (Evaluator, error)
}

type ProblemFactory interface {
	Specification() (*ProblemSpec, error)
	CreateProblem(seed int64) (Problem, error)
}

type ProblemRunner struct {
	factory    ProblemFactory
	problems   map[uint64]Problem
	evaluators map[uint64]Evaluator
}

func NewProblemRunner(factory ProblemFactory) *ProblemRunner {
	return &ProblemRunner{factory, nil, nil}
}

func (r *ProblemRunner) Run() error {
	r.problems = map[uint64]Problem{}
	r.evaluators = map[uint64]Evaluator{}

	if err := r.castProblemSpec(); err != nil {
		return err
	}

	for {
		do_continue, err := r.runOnce()
		if err != nil {
			return err
		}

		if !do_continue {
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
		EvaluatorId uint64 `json:"evaluator_id"`
		NextStep    uint64 `json:"next_step"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	evaluator := r.evaluators[message.EvaluatorId]
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
		EvaluatorId uint64 `json:"evaluator_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	delete(r.evaluators, message.EvaluatorId)
	return nil
}

func (r *ProblemRunner) handleCreateEvaluatorCall(input []byte) error {
	var message struct {
		ProblemId   uint64    `json:"problem_id"`
		EvaluatorId uint64    `json:"evaluator_id"`
		Params      []float64 `json:"params"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	problem := r.problems[message.ProblemId]
	evaluator, err := problem.CreateEvaluator(message.Params)
	if err == UnevalableParamsError {
		return r.sendMessage(map[string]interface{}{"type": "ERROR_REPLY", "kind": "UNEVALABLE_PARAMS"})
	} else if err != nil {
		return err
	}

	r.evaluators[message.EvaluatorId] = evaluator
	return r.sendMessage(map[string]interface{}{"type": "CREATE_EVALUATOR_REPLY"})
}

func (r *ProblemRunner) handleDropProblemCast(input []byte) error {
	var message struct {
		ProblemId uint64 `json:"problem_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	delete(r.problems, message.ProblemId)
	return nil
}

func (r *ProblemRunner) handleCreateProblemCast(input []byte) error {
	var message struct {
		ProblemId  uint64 `json:"problem_id"`
		RandomSeed uint64 `json:"random_seed"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	problem, err := r.factory.CreateProblem(int64(message.RandomSeed))
	if err != nil {
		return err
	}

	r.problems[message.ProblemId] = problem
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
