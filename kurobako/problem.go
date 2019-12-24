package kurobako

type ProblemSpec struct {
	Name   string `json:"name"`
	Attrs  map[string]string `json:"attrs"`
	Params []Var `json:"params_domain"`
	Values []Var `json:"values_domain"`
	Steps  Steps `json:"steps"`
}

type Evaluator interface {
	Evaluate(nextStep uint64) (uint64, []float64, error)
}

type Problem interface {
	CreateEvaluator(params []float64) (*Evaluator, error)
}

type ProblemFactory interface{
	Specification() (ProblemSpec, error)
	CreateProblem(seed int64) (*Problem, error)
}

type ProblemRunner struct {
}
