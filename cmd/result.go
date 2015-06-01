package cmd

import (
	. "github.com/mingzhi/simmlst"
)

type Result struct {
	Ps Config
	C  CovResult
}

type CovResult struct {
	Ks    float64
	KsN   int
	KsVar float64
	Ct    []float64
	CtN   []int
	CtVar []float64
}

type FitResult struct {
	B0, B1, B2                 float64
	Func                       string
	Theta, Rho, Ks             float64
	N, NumGene, LenGene, Delta int
}
