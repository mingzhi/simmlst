package cov

import (
	"github.com/mingzhi/biogo/seq"
)

func subProfile(a, b *seq.Sequence) []float64 {
	var subs []float64
	for i := 0; i < len(a.Seq); i++ {
		var v float64 = 0
		if a.Seq[i] != b.Seq[i] {
			v = 1.0
		}
		subs = append(subs, v)
	}

	return subs
}
