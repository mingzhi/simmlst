package cov

import (
	"github.com/mingzhi/biogo/seq"
	"github.com/mingzhi/gomath/stat/desc/meanvar"
)

type KsCalculator struct {
	*meanvar.MeanVar
}

func (k *KsCalculator) Append(k2 *KsCalculator) {
	k.MeanVar.Append(k2.MeanVar)
}

func NewKsCalculator() *KsCalculator {
	kc := KsCalculator{}
	kc.MeanVar = meanvar.New()
	return &kc
}

func CalcKs(sequences []*seq.Sequence) *KsCalculator {
	ks := NewKsCalculator()
	for i := 0; i < len(sequences); i++ {
		for j := i + 1; j < len(sequences); j++ {
			subs := subProfile(sequences[i], sequences[j])
			for k := 0; k < len(subs); k++ {
				ks.Increment(subs[k])
			}
		}
	}

	return ks
}
