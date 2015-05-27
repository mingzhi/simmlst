package cmd

import (
	. "github.com/mingzhi/simmlst"
)

type Result struct {
	Ps ParameterSet
	C  CovResult
}

type CovResult struct {
	Ks float64
	Ct []float64
}

type FitResult struct {
	Ps         ParameterSet
	B0, B1, B2 float64
	Func       string
}
