package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/gomath/stat/desc/meanvar"
	. "github.com/mingzhi/simmlst"
	. "github.com/mingzhi/simmlst/cmd"
	"os"
	"runtime"
)

var (
	ncpu   int
	input  string
	output string
)

func init() {
	flag.IntVar(&ncpu, "ncpu", runtime.NumCPU(), "ncpu")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)
}

func main() {
	resChan := read(input)
	m := average(resChan)

	resArray := []Result{}
	for ps, a := range m {
		res := Result{}
		res.Ps = ps
		res.C = a.ToCovResult()
		resArray = append(resArray, res)
	}

	write(resArray, output)
}

func read(filename string) chan Result {
	var results []Result
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err := d.Decode(&results); err != nil {
		panic(err)
	}

	resChan := make(chan Result)

	go func() {
		defer close(resChan)
		for _, res := range results {
			resChan <- res
		}
	}()

	return resChan
}

func write(resArray []Result, filename string) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	e := json.NewEncoder(w)
	if err := e.Encode(resArray); err != nil {
		panic(err)
	}
}

func average(resChan chan Result) map[ParameterSet]*averager {
	m := make(map[ParameterSet]*averager)
	for res := range resChan {
		ps := res.Ps
		a, found := m[ps]
		if !found {
			n := len(res.C.Ct)
			a = newAverager(n)
		}
		a.Increment(res.C)
		m[ps] = a
	}

	return m
}

type averager struct {
	Ks *meanvar.MeanVar
	Ct []*meanvar.MeanVar
}

func newAverager(n int) *averager {
	var a averager
	a.Ks = meanvar.New()
	a.Ct = make([]*meanvar.MeanVar, n)
	for i := 0; i < n; i++ {
		a.Ct[i] = meanvar.New()
	}

	return &a
}

func (a *averager) Increment(res CovResult) {
	a.Ks.Increment(res.Ks)
	for i := 0; i < len(res.Ct) && i < len(a.Ct); i++ {
		a.Ct[i].Increment(res.Ct[i])
	}
}

func (a *averager) ToCovResult() CovResult {
	var res CovResult
	res.Ks, res.KsVar, res.KsN = getValuesFromMV(a.Ks)
	for i := 0; i < len(a.Ct); i++ {
		m, v, n := getValuesFromMV(a.Ct[i])
		res.Ct = append(res.Ct, m)
		res.CtVar = append(res.CtVar, v)
		res.CtN = append(res.CtN, n)
	}

	return res
}

func getValuesFromMV(mv *meanvar.MeanVar) (m, v float64, n int) {
	m = mv.Mean.GetResult()
	v = mv.Var.GetResult()
	n = mv.Mean.GetN()
	return
}
