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
