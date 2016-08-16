package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/mingzhi/biogo/seq"
	"github.com/mingzhi/simmlst"
	"io/ioutil"
	"math"
	"os"
	"runtime"
)

var (
	cfgFile = kingpin.Arg("cfg", "population configure file").String()
	outFile = kingpin.Arg("out", "out file").String()
	ncpu    = kingpin.Flag("ncpu", "number of CPUs").Default("1").Int()
	maxl    = kingpin.Flag("maxl", "max length of correlation").Default("100").Int()
	repeat  = kingpin.Flag("repeat", "repeat").Default("1").Int()
)

func main() {
	kingpin.Parse()
	runtime.GOMAXPROCS(*ncpu)

	cfg := readCfg(*cfgFile)

	jobChan := make(chan simmlst.Config)
	go func() {
		defer close(jobChan)
		for k := 0; k < *repeat; k++ {
			jobChan <- cfg
		}
	}()

	resChan := make(chan Result)
	done := make(chan bool)
	for k := 0; k < *ncpu; k++ {
		go func() {
			for c := range jobChan {
				rc := runSimmlst(c)
				for r := range rc {
					resChan <- r
				}
			}
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for k := 0; k < *ncpu; k++ {
			<-done
		}
	}()

	res := collect(resChan, *maxl)
	write(res, *outFile)
}

// runSimmlst executes simmlst.
func runSimmlst(cfg simmlst.Config) chan Result {
	resChan := make(chan Result)
	go func() {
		defer close(resChan)
		// create tmp file.
		tmp, _ := ioutil.TempFile("", "simmlst")
		defer os.Remove(tmp.Name())
		// execute simmlst.
		simmlst.Exec(cfg, tmp.Name())
		// collect simulation results and calculate correlations.
		alignments := seq.ReadXMFA(tmp.Name())
		for _, a := range alignments {
			genes := []string{}
			for _, g := range a {
				genes = append(genes, string(g.Seq))
			}
			results := calcCorr(genes, *maxl)
			for _, r := range results {
				resChan <- r
			}
		}
	}()

	return resChan
}

// readCfg read and return a population configuration.
func readCfg(file string) simmlst.Config {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var c simmlst.Config
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&c); err != nil {
		panic(err)
	}
	return c
}

// calcCorr calculates correlation functions from an alignment.
func calcCorr(alignment []string, maxl int) (results []Result) {
	cms := calcCmSub(alignment, maxl)
	results = append(results, cms...)

	return
}

// collect averages correlation results.
func collect(resChan chan Result, maxLen int) map[string][]*MeanVar {
	resMap := make(map[string][]*MeanVar)
	for res := range resChan {
		for len(resMap[res.Type]) <= res.Lag {
			resMap[res.Type] = append(resMap[res.Type], NewMeanVar())
		}
		if !math.IsNaN(res.Value) {
			resMap[res.Type][res.Lag].Add(res.Value)
		}
	}

	return resMap
}

// write the final result.
func write(result map[string][]*MeanVar, outFile string) {
	w, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("l,m,v,n,t\n")
	for t, mvs := range result {
		for i := 0; i < len(mvs); i++ {
			m := mvs[i].Mean()
			v := mvs[i].Variance()
			n := mvs[i].N
			if n > 0 && !math.IsNaN(v) {
				w.WriteString(fmt.Sprintf("%d", i))
				w.WriteString(fmt.Sprintf(",%g,%g", m, v))
				w.WriteString(fmt.Sprintf(",%d,%s\n", n, t))
			}
		}
	}
}
