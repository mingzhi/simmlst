package cov

import (
	"github.com/mingzhi/biogo/seq"
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
)

type CovCalculatorFFT struct {
	N            int
	maskXYs, xys []float64
	mean         *desc.Mean
	circular     bool
}

func NewCovCalculatorFFT(maxl int, circular bool) *CovCalculatorFFT {
	var c CovCalculatorFFT
	c.N = maxl
	c.maskXYs = make([]float64, c.N)
	c.xys = make([]float64, c.N)
	c.mean = desc.NewMean()
	c.circular = circular

	return &c
}

func (c *CovCalculatorFFT) Increment(xs []float64) {
	masks := make([]float64, len(xs))
	for i := 0; i < len(masks); i++ {
		masks[i] = 1.0
		c.mean.Increment(xs[i])
	}

	maskXYs := correlation.AutoCorrFFT(masks, c.circular)
	xys := correlation.AutoCorrFFT(xs, c.circular)

	for i := 0; i < len(c.xys); i++ {
		c.xys[i] += (xys[i] + xys[(len(xys)-i)%len(xys)])
		c.maskXYs[i] += (maskXYs[i] + maskXYs[(len(maskXYs)-i)%len(maskXYs)])
	}
}

func (c *CovCalculatorFFT) Append(c2 *CovCalculatorFFT) {
	c.mean.Append(c2.mean)
	for i := 0; i < len(c2.xys); i++ {
		c.xys[i] += c2.xys[i]
		c.maskXYs[i] += c2.maskXYs[i]
	}
}

func (c *CovCalculatorFFT) GetResult(i int) float64 {
	pxy := c.xys[i] / c.maskXYs[i]
	pxpy := c.mean.GetResult() * c.mean.GetResult()
	return pxy - pxpy
}

type CovCalculator struct {
	N     int
	corrs []*correlation.BivariateCovariance
}

func NewCovCalculator(maxl int, bias bool) *CovCalculator {
	cc := CovCalculator{}
	cc.corrs = make([]*correlation.BivariateCovariance, maxl)
	for i := 0; i < maxl; i++ {
		cc.corrs[i] = correlation.NewBivariateCovariance(bias)
	}
	cc.N = maxl
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

func CalcCtFFT(sequences []*seq.Sequence, maxl int) *CovCalculatorFFT {
	ct := NewCovCalculatorFFT(maxl, false)
	for i := 0; i < len(sequences); i++ {
		for j := i + 1; j < len(sequences); j++ {
			subs := subProfile(sequences[i], sequences[j])
			ct.Increment(subs)
		}
	}

	return ct
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
