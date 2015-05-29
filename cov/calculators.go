package cov

type Calculators struct {
	Ks *KsCalculator
	Ct *CovCalculator
}

func NewCalculators(maxl int, bias bool) *Calculators {
	var c Calculators
	c.Ks = NewKsCalculator()
	c.Ct = NewCovCalculator(maxl, bias)
	return &c
}

func (c *Calculators) Append(c2 *Calculators) {
	c.Ct.Append(c2.Ct)
	c.Ks.Append(c2.Ks)
}

type CalculatorsFFT struct {
	Ks *KsCalculator
	Ct *CovCalculatorFFT
}

func NewCalculatorsFFT(maxl int, circular bool) *CalculatorsFFT {
	var c CalculatorsFFT
	c.Ks = NewKsCalculator()
	c.Ct = NewCovCalculatorFFT(maxl, circular)
	return &c
}

func (c *CalculatorsFFT) Append(c2 *CalculatorsFFT) {
	c.Ct.Append(c2.Ct)
	c.Ks.Append(c2.Ks)
}
