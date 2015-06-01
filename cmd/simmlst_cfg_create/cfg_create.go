package main

import (
	"encoding/json"
	"flag"
	"github.com/mingzhi/simmlst"
	"os"
)

func main() {
	var input, output string
	var rep int
	var par ParameterSet
	var cfgs []simmlst.Config
	flag.IntVar(&rep, "r", 1, "replicates")
	flag.Parse()
	input = flag.Arg(0)
	output = flag.Arg(1)

	par = parse(input)
	cfgs = create(par, rep)
	write(output, cfgs)
}

type ParameterSet struct {
	Sizes    []int
	NumGenes []int
	LenGenes []int
	Thetas   []float64
	Rhos     []float64
	Deltas   []int
}

func create(par ParameterSet, rep int) []simmlst.Config {
	var cfgs []simmlst.Config
	for _, size := range par.Sizes {
		for _, numGene := range par.NumGenes {
			for _, lenGene := range par.LenGenes {
				for _, theta := range par.Thetas {
					for _, rho := range par.Rhos {
						for _, deltas := range par.Deltas {
							for i := 0; i < rep; i++ {
								var cfg simmlst.Config
								cfg.N = size
								cfg.NumGene = numGene
								cfg.LenGene = lenGene
								cfg.Theta = theta
								cfg.Rho = rho
								cfg.Delta = deltas
								cfgs = append(cfgs, cfg)
							}
						}
					}
				}
			}
		}
	}

	return cfgs
}

func parse(filename string) ParameterSet {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var sets ParameterSet

	d := json.NewDecoder(f)
	if err := d.Decode(&sets); err != nil {
		panic(err)
	}

	return sets
}

func write(filename string, configs []simmlst.Config) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	e := json.NewEncoder(w)
	if err := e.Encode(configs); err != nil {
		panic(err)
	}
}
