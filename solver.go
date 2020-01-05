package kurobako

import (
	"encoding/json"
	"fmt"
	"io"
)

// SolverSpec is the specification of a solver.
type SolverSpec struct {
	// Name is the name of the solver.
	Name         string            `json:"name"`

	// Attrs is the attributes of the solver.
	Attrs        map[string]string `json:"attrs"`

	// Capabilities is the capabilities of the solver.
	Capabilities Capabilities      `json:"capabilities"`
}

// NewSolverSpec creates a new SolverSpec instance.
func NewSolverSpec(name string) SolverSpec {
	return SolverSpec{name, map[string]string{}, AllCapabilities}
}

// Solver interface.
type Solver interface {
	// Ask returns a NextTrial object that contains information about the next trial to be evaluated.
	Ask(idg *TrialIDGenerator) (NextTrial, error)

	// Tell takes an evaluation result of a trial and updates the state of the solver.
	Tell(trial EvaluatedTrial) error
}

// SolverFactory allows to create a new solver instance.
type SolverFactory interface {
	// Specification returns the specification of the solver.
	Specification() (*SolverSpec, error)

	// CreateSolver creates a new solver instance with the given random seed.
	//
	// The created solver will be used to solve a black-box optimization problem defined by the given ProblemSpec.
	CreateSolver(seed int64, problem ProblemSpec) (Solver, error)
}

// SolverRunner runs a solver.
type SolverRunner struct {
	factory SolverFactory
	solvers map[uint64]Solver
}

// NewSolverRunner creates a new SolverRunner instance that handles the given solver.
func NewSolverRunner(factory SolverFactory) *SolverRunner {
	return &SolverRunner{factory, nil}
}

// Run runs a solver.
func (r *SolverRunner) Run() error {
	r.solvers = map[uint64]Solver{}

	if err := r.castSolverSpec(); err != nil {
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

func (r *SolverRunner) runOnce() (bool, error) {
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
	case "CREATE_SOLVER_CAST":
		err := r.handleCreateSolverCast(line)
		return true, err
	case "DROP_SOLVER_CAST":
		err := r.handleDropSolverCast(line)
		return true, err
	case "ASK_CALL":
		err := r.handleAskCall(line)
		return true, err
	case "TELL_CALL":
		err := r.handleTellCall(line)
		return true, err
	default:
		return false, fmt.Errorf("unknown message type: %v", message["type"])
	}
}

func (r *SolverRunner) handleTellCall(input []byte) error {
	var message struct {
		SolverID uint64         `json:"solver_id"`
		Trial    EvaluatedTrial `json:"trial"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	solver := r.solvers[message.SolverID]
	if err := solver.Tell(message.Trial); err != nil {
		return err
	}

	reply := map[string]interface{}{
		"type": "TELL_REPLY",
	}
	return r.sendMessage(reply)
}

func (r *SolverRunner) handleAskCall(input []byte) error {
	var message struct {
		SolverID    uint64 `json:"solver_id"`
		NextTrialID uint64 `json:"next_trial_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	idg := TrialIDGenerator{message.NextTrialID}
	solver := r.solvers[message.SolverID]
	trial, err := solver.Ask(&idg)
	if err != nil {
		return err
	}

	reply := map[string]interface{}{
		"type":          "ASK_REPLY",
		"trial":         trial,
		"next_trial_id": idg.NextID,
	}
	return r.sendMessage(reply)
}

func (r *SolverRunner) handleDropSolverCast(input []byte) error {
	var message struct {
		SolverID uint64 `json:"solver_id"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	delete(r.solvers, message.SolverID)
	return nil
}

func (r *SolverRunner) handleCreateSolverCast(input []byte) error {
	var message struct {
		SolverID   uint64      `json:"solver_id"`
		RandomSeed uint64      `json:"random_seed"`
		Problem    ProblemSpec `json:"problem"`
	}

	if err := json.Unmarshal(input, &message); err != nil {
		return err
	}

	solver, err := r.factory.CreateSolver(int64(message.RandomSeed), message.Problem)
	if err != nil {
		return err
	}

	r.solvers[message.SolverID] = solver
	return nil
}

func (r *SolverRunner) castSolverSpec() error {
	spec, err := r.factory.Specification()
	if err != nil {
		return err
	}

	return r.sendMessage(map[string]interface{}{"type": "SOLVER_SPEC_CAST", "spec": spec})
}

func (r *SolverRunner) sendMessage(message map[string]interface{}) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(bytes))
	return nil
}
