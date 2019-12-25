package main

import (
	"github.com/c-bata/goptuna"
	"github.com/sile/kurobako-go"
	"github.com/sile/kurobako-go/goptuna/solver"
)

func createStudy(seed int64) (*goptuna.Study, error) {
	return goptuna.CreateStudy("example-study")
}

func main() {
	factory := solver.NewGoptunaSolverFactory(createStudy)
	runner := kurobako.NewSolverRunner(&factory)
	if err := runner.Run(); err != nil {
		panic(err)
	}
}
