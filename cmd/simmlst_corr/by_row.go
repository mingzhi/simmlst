package main

import (
	"bitbucket.org/mingzhi/seqcorr/nuclcov"
	"math"
	"sort"
)

type Result struct {
	Lag   int
	Value float64
	N     int
	Type  string
}

// Sub records the position and
type Sub struct {
	Pos int
	A   byte
}

// Subs is a list of Sub.
type Subs []Sub

// Len returns the length of Subs.
func (s Subs) Len() int { return len(s) }

// Swap swap the values at two positions.
func (s Subs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// BySubPos is a wrapper for sorting.
type BySubPos struct{ Subs }

// Less return true if the value at i is less than the value at j.
func (s BySubPos) Less(i, j int) bool {
	return s.Subs[i].Pos < s.Subs[j].Pos
}

func calcCs(genomes []string, maxl int) (results []Result) {
	matrix := [][]*nuclcov.NuclCov{}
	for _, genome := range genomes {
		for i := 0; i < len(genome); i++ {
			for j := i; j < len(genome) && j-i < maxl; j++ {
				pos := i
				lag := j - i
				a := genome[i]
				b := genome[j]
				for len(matrix) <= pos {
					matrix = append(matrix, []*nuclcov.NuclCov{})
				}

				for len(matrix[pos]) <= lag {
					matrix[pos] = append(matrix[pos], nuclcov.New([]byte{'1', '2', '3', '4'}))
				}

				matrix[pos][lag].Add(a, b)
			}
		}
	}

	for lag := 0; lag < maxl; lag++ {
		mc := NewMeanCov()
		for i := 0; i < len(matrix); i++ {
			if lag < len(matrix[i]) {
				xy, xbar, ybar, n := matrix[i][lag].Cov()
				if !math.IsNaN(xy) {
					mc.Add(xy, xbar, ybar, n)
				}
			}
		}

		cs := mc.Mean.GetResult()
		cr := mc.Cov.GetResult()
		p2 := mc.MeanXY()
		n := mc.Mean.GetN()

		crRes := Result{Value: cr, Lag: lag, N: n, Type: "Cr"}
		csRes := Result{Value: cs, Lag: lag, N: n, Type: "Cs"}
		p2Res := Result{Value: p2, Lag: lag, N: n, Type: "P2"}

		results = append(results, []Result{crRes, csRes, p2Res}...)
	}

	return
}

func calcCm(genomes []string, maxl int) (results []Result) {
	length := len(genomes[0])
	subBuf := make([]float64, length)
	cms := make([]float64, maxl)
	d := 0.0
	vd := 0.0

	for i := 0; i < len(genomes); i++ {
		for j := i + 1; j < len(genomes); j++ {
			for k := 0; k < length; k++ {
				if genomes[i][k] != genomes[j][k] {
					subBuf[k] = 1.0
				} else {
					subBuf[k] = 0.0
				}
			}

			var xbar, ybar float64
			for l := 0; l < maxl; l++ {
				xy := 0.0
				for k := 0; k < length; k++ {
					xy += subBuf[k] * subBuf[(k+l)%length]
				}

				xy /= float64(length)
				if l == 0 {
					xbar = xy
					ybar = xy
					d += xbar
					vd += xbar * ybar
				}

				cms[l] += xy - xbar*ybar
			}
		}
	}

	n := len(genomes) * (len(genomes) - 1) / 2
	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "Cm"
		res.Value = cms[i] / float64(n)
		results = append(results, res)
	}

	ks := d / float64(n)
	vard := vd/float64(n) - ks*ks

	results = append(results, Result{Lag: 0, N: n, Type: "Ks", Value: ks})
	results = append(results, Result{Lag: 0, N: n, Type: "Vd", Value: vard})

	return
}

func calcCmSub(genomes []string, maxl int) (results []Result) {
	subsArr := identifySubs(genomes)
	length := len(genomes[0])

	totals := make([]float64, maxl)
	d := 0.0
	vd := 0.0
	for i := 0; i < len(subsArr); i++ {
		for j := i + 1; j < len(subsArr); j++ {
			allSubs := removeDuplicateSubs(subsArr[i], subsArr[j])
			positions := []int{}
			for _, s := range allSubs {
				positions = append(positions, s.Pos)
			}

			xy := make([]int, maxl)
			for k := 0; k < len(positions); k++ {
				for h := 0; h < len(positions); h++ {
					lag := (positions[h] - positions[k] + length) % length
					if lag < len(xy) {
						xy[lag]++
					}
				}
			}

			totalSubs := len(positions)
			xbar := float64(totalSubs) / float64(length)
			ybar := xbar
			xbarybar := xbar * ybar
			d += xbar
			vd += xbarybar

			for lag := 0; lag < maxl; lag++ {
				v := float64(xy[lag])/float64(length) - xbarybar
				totals[lag] += v
			}
		}
	}

	n := len(subsArr) * (len(subsArr) - 1) / 2

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "Cm"
		res.Value = totals[i] / float64(n)
		results = append(results, res)
	}

	ks := d / float64(n)
	vard := vd/float64(n) - ks*ks

	results = append(results, Result{Lag: 0, N: n, Type: "Ks", Value: ks})
	results = append(results, Result{Lag: 0, N: n, Type: "Vd", Value: vard})

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "Cm2"
		res.Value = totals[i] / (float64(n) * ks)
		results = append(results, res)
	}

	return
}

func identifySubs(genomes []string) (subsArr []Subs) {
	ref := genomes[0]
	subsArr = append(subsArr, Subs{})
	for i := 1; i < len(genomes); i++ {
		subs := Subs{}
		for k := 0; k < len(ref); k++ {
			if ref[k] != genomes[i][k] {
				subs = append(subs, Sub{Pos: k, A: genomes[i][k]})
			}
		}
		subsArr = append(subsArr, subs)
	}

	return
}

// removeDuplicateSubs
func removeDuplicateSubs(subs1 Subs, others ...Subs) Subs {
	allSubs := Subs{}
	allSubs = append(allSubs, subs1...)
	for _, subs := range others {
		allSubs = append(allSubs, subs...)
	}
	if len(allSubs) <= 1 {
		return allSubs
	}
	// remove same subsistutions.
	sort.Sort(BySubPos{allSubs})

	dedupSubs := Subs{}
	old := Sub{Pos: allSubs[0].Pos - 1, A: ' '}
	for _, s := range allSubs {
		if s.Pos == old.Pos {
			if s.A == old.A {
				dedupSubs = dedupSubs[:len(dedupSubs)-1]
			}
		} else {
			dedupSubs = append(dedupSubs, s)
		}
		old = s
	}

	return dedupSubs
}

// removeDuplicateInts
func removeDuplicateInts(values []int, others ...[]int) []int {
	all := []int{}
	others = append(others, values)
	for _, vs := range others {
		all = append(all, vs...)
	}

	if len(all) <= 1 {
		return all
	}

	sort.Ints(all)
	dedupInts := []int{}
	old := all[0] - 1
	for _, s := range all {
		if s == old {
			dedupInts = dedupInts[:len(dedupInts)-1]
		} else {
			dedupInts = append(dedupInts, s)
		}
		old = s
	}

	return dedupInts
}
