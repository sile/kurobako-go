package kurobako

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"io"
)

type ProblemSpec struct {
	Name   string            `json:"name"`
	Attrs  map[string]string `json:"attrs"`
	Params []Var             `json:"params_domain"`
	Values []Var             `json:"values_domain"`
	Steps  Steps             `json:"steps"`
}

type Evaluator interface {
	Evaluate(nextStep uint64) (currentStep uint64, values []float64, err error)
}

type Problem interface {
	CreateEvaluator(params []float64) (Evaluator, error)
}

type ProblemFactory interface {
	Specification() (ProblemSpec, error)
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
			return nil
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
		panic("todo")
	case "DROP_PROBLEM_CAST":
		panic("todo")
	case "CREATE_EVALUATOR_CALL":
		panic("todo")
	case "DROP_EVALUATOR_CALL":
		panic("todo")
	case "EVALUATE_CALL":
		panic("todo")
	default:
		return false, fmt.Errorf("unknown message type: %v", message["type"])
	}

	return true, nil
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
