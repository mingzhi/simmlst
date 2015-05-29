package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/meta/fit"
	. "github.com/mingzhi/simmlst/cmd"
	"math"
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
	fitResults := batchFit(resChan)
	write(fitResults, output)
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

func write(fitResults []FitResult, filename string) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	e := json.NewEncoder(w)
	if err := e.Encode(fitResults); err != nil {
		panic(err)
	}
}

type FitControl struct {
	FitFunc    FitFunc
	Start, End int
}

func batchFit(resChan chan Result) []FitResult {
	fitFuncMap := make(map[string]FitControl)
	fitFuncMap["Exp"] = FitControl{FitFunc: fit.FitExp, Start: 1, End: -1}
	fitFuncMap["Hyper"] = FitControl{FitFunc: fit.FitHyper, Start: 1, End: 10}

	numWorker := ncpu
	done := make(chan bool)
	fitResChan := make(chan FitResult)
	worker := func() {
		defer send(done)
		for res := range resChan {
			for name, fitControl := range fitFuncMap {
				fitFunc := fitControl.FitFunc
				start := fitControl.Start
				end := fitControl.End
				if end < 0 {
					end = len(res.C.Ct)
				}
				fitRes := doFit(res, fitFunc, start, end)
				fitRes.Func = name
				fitResChan <- fitRes
			}
		}
	}

	for i := 0; i < numWorker; i++ {
		go worker()
	}

	go func() {
		defer close(fitResChan)
		for i := 0; i < numWorker; i++ {
			<-done
		}
	}()

	fitResults := []FitResult{}
	for fitRes := range fitResChan {
		fitResults = append(fitResults, fitRes)
	}

	return fitResults
}

func send(done chan bool) {
	done <- true
}

type FitFunc func(xdata, ydata []float64) (par []float64)

func doFit(res Result, fitFunc FitFunc, start, end int) FitResult {
	var fitRes FitResult
	var xdata, ydata []float64
	for i := start; i < end; i++ {
		v := res.C.Ct[i]
		if !math.IsNaN(v) {
			xdata = append(xdata, float64(i))
			ydata = append(ydata, v)
		}
	}

	par := fitFunc(xdata, ydata)
	fitRes.B0 = par[0]
	fitRes.B1 = par[1]
	if len(par) > 2 {
		fitRes.B2 = par[2]
	}

	fitRes.Delta = res.Ps.Delta
	fitRes.LenGene = res.Ps.LenGene
	fitRes.N = res.Ps.N
	fitRes.NumGene = res.Ps.NumGene
	fitRes.Rho = res.Ps.Rho
	fitRes.Theta = res.Ps.Theta
	fitRes.Ks = res.C.Ks

	return fitRes
}
