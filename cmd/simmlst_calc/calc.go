// Run SimMLST simulator, and fit correlation functions.
package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/biogo/seq"
	. "github.com/mingzhi/simmlst"
	. "github.com/mingzhi/simmlst/cmd"
	. "github.com/mingzhi/simmlst/cov"
	. "github.com/mingzhi/simmlst/io"
	"io"
	"io/ioutil"
	"os"
	"runtime"
)

var (
	maxl   int
	ncpu   int
	input  string
	output string
)

func init() {
	flag.IntVar(&maxl, "maxl", 200, "maxl")
	flag.IntVar(&ncpu, "ncpu", runtime.NumCPU(), "ncpu")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)

	runtime.GOMAXPROCS(ncpu)
}

func main() {
	psChan := read(input)
	resChan := runSimulation(psChan)
	results := collect(resChan)
	write(output, results)
}

func read(filename string) chan ParameterSet {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	psArr := readParameterSets(f)
	return streamPS(psArr)
}

func collect(resChan chan tempResult) []Result {
	m := make(map[ParameterSet]*Calculators)
	for res := range resChan {
		c, found := m[res.Ps]
		if !found {
			c = res.C
		} else {
			c.Append(res.C)
		}
		m[res.Ps] = c
	}

	var results []Result
	for ps, c := range m {
		res := Result{}
		res.Ps = ps
		res.C = createCovResult(c)
		results = append(results)
	}

	return results
}

func write(filename string, results []Result) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(results); err != nil {
		panic(err)
	}
}

func createCovResult(c *Calculators) CovResult {
	var cr CovResult
	cr.Ks = c.Ks.Mean.GetResult()
	for i := 0; i < c.Ct.N; i++ {
		cr.Ct = append(cr.Ct, c.Ct.GetResult(i))
	}
	return cr
}

type tempResult struct {
	Ps ParameterSet
	C  *Calculators
}

func runSimulation(psChan chan ParameterSet) chan tempResult {
	ncpu := runtime.GOMAXPROCS(0)
	numWorker := ncpu

	resChan := make(chan tempResult)
	done := make(chan bool)

	worker := func() {
		defer send(done)
		for ps := range psChan {
			tempfile, _ := ioutil.TempFile("", "simmlst")

			Exec(ps, tempfile.Name())

			geneGroups := readSequences(tempfile.Name())
			c := calcCorr(geneGroups, maxl)

			resChan <- tempResult{Ps: ps, C: c}

			os.Remove(tempfile.Name())
		}
	}

	for i := 0; i < numWorker; i++ {
		go worker()
	}

	go func() {
		defer close(resChan)
		wait(done, numWorker)
	}()

	return resChan
}

func readSequences(filename string) (geneGroups [][]*seq.Sequence) {
	geneGroups = ReadXMFA(filename)
	return
}

func calcCorr(geneGroups [][]*seq.Sequence, maxl int) (c *Calculators) {
	for i := 0; i < len(geneGroups); i++ {
		c1 := calcCorrOne(geneGroups[i], maxl)
		if i == 0 {
			c = c1
		} else {
			c.Ct.Append(c1.Ct)
			c.Ks.Append(c1.Ks)
		}
	}
	return
}

func calcCorrOne(genes []*seq.Sequence, maxl int) *Calculators {
	var c Calculators
	c.Ct = CalcCt(genes, maxl)
	c.Ks = CalcKs(genes)
	return &c
}

func streamPS(psArr []ParameterSet) chan ParameterSet {
	c := make(chan ParameterSet)
	go func() {
		defer close(c)
		for _, ps := range psArr {
			c <- ps
		}
	}()
	return c
}

func readParameterSets(r io.Reader) []ParameterSet {
	var psArr []ParameterSet
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&psArr)
	if err != nil {
		panic(err)
	}
	return psArr
}

func send(done chan bool) {
	done <- true
}

func wait(done chan bool, numWorker int) {
	for i := 0; i < numWorker; i++ {
		<-done
	}
}
