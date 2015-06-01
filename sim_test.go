package simmlst

import (
	"github.com/mingzhi/gomath/stat/desc/meanvar"
	"github.com/mingzhi/meta/fit"
	. "github.com/mingzhi/simmlst/cov"
	. "github.com/mingzhi/simmlst/io"
	"math"
	"testing"
)

func TestExec(t *testing.T) {
	var ps Config
	ps.N = 10
	ps.Delta = 50
	ps.LenGene = 10000
	ps.NumGene = 1
	ps.Theta = 1000
	ps.Rho = 1000
	ps.DataFile = "test.xmfa"

	mv := meanvar.New()
	mv2 := meanvar.New()
	maxl := 100

	n := 10
	for i := 0; i < n; i++ {
		Exec(ps)

		geneGroups := ReadXMFA(ps.DataFile)

		// ksCalculators := []*KsCalculator{}
		ctCalculators := []*CovCalculator{}
		for i := 0; i < len(geneGroups); i++ {
			// ksCalculators = append(ksCalculators, CalcKs(geneGroups[i]))
			ctCalculators = append(ctCalculators, CalcCt(geneGroups[i], maxl))
		}

		// ks := NewKsCalculator()
		ct := NewCovCalculator(maxl, false)
		for i := 0; i < len(ctCalculators); i++ {
			// ks.Append(ksCalculators[i])
			ct.Append(ctCalculators[i])
		}

		xdata := []float64{}
		ydata := []float64{}
		for i := 1; i < maxl; i++ {
			xdata = append(xdata, float64(i))
			ydata = append(ydata, ct.GetResult(i))
		}

		par := fit.FitExp(xdata, ydata)
		par2 := fit.FitHyper(xdata[:10], ydata[:10])
		rm := par2[1] * 0.5 / (math.Sqrt(par2[0]))

		mv.Increment(rm)
		mv2.Increment(par[2])
	}

	// theta := ps.Theta / float64(2*ps.LenGene*ps.NumGene)
	expectedRm := ps.Rho / ps.Theta
	resultedRm := mv.Mean.GetResult()
	stdError := math.Sqrt(mv.Var.GetResult() / float64(mv.Var.GetN()))
	if math.Abs(expectedRm-resultedRm) > 1.96*stdError {
		t.Errorf("Expect rm %f, got %f, at std error %f\n", expectedRm, resultedRm, stdError)
	}

	// theta := ps.Theta / float64(2*ps.LenGene*ps.NumGene)
	expectedDelta := float64(ps.Delta)
	resultedDelta := mv2.Mean.GetResult()
	stdErrorDelta := math.Sqrt(mv2.Var.GetResult() / float64(mv2.Var.GetN()))
	if math.Abs(expectedDelta-resultedDelta) > 1.96*stdErrorDelta {
		t.Errorf("Expect delta %f, got %f, at std error %f\n", expectedDelta, resultedDelta, stdErrorDelta)
	}

}
