// Run SimMLST simulator, and fit correlation functions.
package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/biogo/seq"
	"github.com/mingzhi/gomath/stat/correlation"
	"github.com/mingzhi/seqcor/calculator"
	. "github.com/mingzhi/simmlst"
	. "github.com/mingzhi/simmlst/cmd"
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
	flag.IntVar(&maxl, "maxl", 1000, "maxl")
	flag.IntVar(&ncpu, "ncpu", runtime.NumCPU(), "ncpu")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)

	runtime.GOMAXPROCS(ncpu)
}

func main() {
	psChan := read(input)
	psMap := make(map[int][]Config)
	for ps := range psChan {
		psMap[ps.LenGene] = append(psMap[ps.LenGene], ps)
	}
	var results []Result
	for seqLen, psSet := range psMap {
		psChan := make(chan Config)
		go func() {
			defer close(psChan)
			for _, ps := range psSet {
				psChan <- ps
			}
		}()
		resChan := run(psChan, seqLen)
		res := collect(resChan)
		results = append(results, res...)
	}
	write(output, results)
}

func read(filename string) chan Config {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	psArr := readConfigs(f)
	return streamPS(psArr)
}

func collect(resChan chan tempResult) []Result {
	m := make(map[Config]*calculators)
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
		res.C = createCovResult(c, maxl)
		results = append(results, res)
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

type calculators struct {
	Ks *calculator.Ks
	Ct *calculator.AutoCovFFTW
}

func (c *calculators) Append(c2 *calculators) {
	c.Ks.Append(c2.Ks)
	c.Ct.Append(c2.Ct)
}

func createCovResult(c *calculators, maxl int) CovResult {
	var cr CovResult
	cr.Ks = c.Ks.Mean.GetResult()
	for i := 0; i < c.Ct.N && i < maxl; i++ {
		cr.Ct = append(cr.Ct, c.Ct.GetResult(i))
	}
	return cr
}

type tempResult struct {
	Ps Config
	C  *calculators
}

func run(psChan chan Config, seqLen int) chan tempResult {
	ncpu := runtime.GOMAXPROCS(0)
	numWorker := ncpu

	circular := false
	dft := correlation.NewFFTW(seqLen, circular)

	resChan := make(chan tempResult)
	done := make(chan bool)

	worker := func() {
		defer send(done)
		for ps := range psChan {
			tempfile, _ := ioutil.TempFile("", "simmlst")

			Exec(ps, tempfile.Name())

			geneGroups := readSequences(tempfile.Name())
			c := calcCorr(geneGroups, &dft)

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

func calcCorr(geneGroups [][]*seq.Sequence, dft *correlation.FFTW) (c *calculators) {
	for i := 0; i < len(geneGroups); i++ {
		c1 := calcCorrOne(geneGroups[i], dft)
		if i == 0 {
			c = c1
		} else {
			c.Ct.Append(c1.Ct)
			c.Ks.Append(c1.Ks)
		}
	}
	return
}

func calcCorrOne(genes []*seq.Sequence, dft *correlation.FFTW) *calculators {
	var c calculators
	var sequences [][]byte
	for _, g := range genes {
		sequences = append(sequences, g.Seq)
	}
	c.Ct = calculator.CalcCtFFTW(sequences, dft)
	c.Ks = calculator.CalcKs(sequences)
	return &c
}

func streamPS(psArr []Config) chan Config {
	c := make(chan Config)
	go func() {
		defer close(c)
		for _, ps := range psArr {
			c <- ps
		}
	}()
	return c
}

func readConfigs(r io.Reader) []Config {
	var psArr []Config
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
