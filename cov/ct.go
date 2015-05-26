package cov

import (
	"github.com/mingzhi/biogo/seq"
	"github.com/mingzhi/gomath/stat/correlation"
)

type CovCalculator struct {
	corrs []*correlation.BivariateCovariance
}

func NewCovCalculator(maxl int, bias bool) *CovCalculator {
	cc := CovCalculator{}
	cc.corrs = make([]*correlation.BivariateCovariance, maxl)
	for i := 0; i < maxl; i++ {
		cc.corrs[i] = correlation.NewBivariateCovariance(bias)
	}
	return &cc
}

func (cc *CovCalculator) Increment(i int, x, y float64) {
	cc.corrs[i].Increment(x, y)
}

func (cc *CovCalculator) GetResult(i int) float64 {
	return cc.corrs[i].GetResult()
}

func (cc *CovCalculator) GetMeanXY(i int) float64 {
	return cc.corrs[i].MeanX() * cc.corrs[i].MeanY()
}

func (cc *CovCalculator) GetN(i int) int {
	return cc.corrs[i].GetN()
}

func (cc *CovCalculator) Append(cc2 *CovCalculator) {
	for i := 0; i < len(cc.corrs); i++ {
		cc.corrs[i].Append(cc2.corrs[i])
	}
}

func CalcCt(sequences []*seq.Sequence, maxl int) *CovCalculator {
	ct := NewCovCalculator(maxl, false)
	for i := 0; i < len(sequences); i++ {
		for j := i + 1; j < len(sequences); j++ {
			subs := subProfile(sequences[i], sequences[j])

			for l := 0; l < maxl; l++ {
				for k := 0; k < len(subs)-l; k++ {
					x, y := subs[k], subs[k+l]
					ct.Increment(l, x, y)
				}
			}
		}
	}

	return ct
}
