package main

import (
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/gomath/stat/desc"
)

// MeanCov implements the law of total covariance.
type MeanCov struct {
	xy, xbar, ybar float64
	n              int
	Mean           *desc.Mean
	Cov            *correlation.BivariateCovariance
}

// NewMeanCov returns a new MeanCov.
func NewMeanCov() *MeanCov {
	mc := &MeanCov{}
	mc.Mean = desc.NewMean()
	mc.Cov = correlation.NewBivariateCovariance(false)
	return mc
}

// Add adds data.
func (m *MeanCov) Add(xy, xbar, ybar float64, n int) {
	m.xy += xy * float64(n)
	m.xbar += xbar * float64(n)
	m.ybar += ybar * float64(n)
	m.n += n
	m.Mean.Increment(xy - xbar*ybar)
	m.Cov.Increment(xbar, ybar)
}

// Ct returns the total covariance.
func (m *MeanCov) Ct() float64 {
	return m.xy/float64(m.n) - m.xbar/float64(m.n)*m.ybar/float64(m.n)
}

// N returns the number of rows.
func (m *MeanCov) N() int {
	return m.n
}

// MeanXY returns the mean of XY.
func (m *MeanCov) MeanXY() float64 {
	return m.xy / float64(m.n)
}

// MeanXbar returns the mean of xbar
func (m *MeanCov) MeanXbar() float64 {
	return m.xbar / float64(m.n)
}

// MeanYbar returns the mean of ybar
func (m *MeanCov) MeanYbar() float64 {
	return m.ybar / float64(m.n)
}
