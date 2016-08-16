package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/mingzhi/simmlst"
	"os"
)

var (
	cfgFile  = kingpin.Arg("cfg", "configs file").String()
	prefix   = kingpin.Flag("prefix", "output prefix").Default("test").String()
	ppn      = kingpin.Flag("ppn", "ppn").Default("1").Int()
	walltime = kingpin.Flag("walltime", "walltime").Default("48").Int()
	message  = kingpin.Flag("message", "message").Default("a").String()
	exec     = kingpin.Flag("exec", "exec name").Default("simmlst_corr").String()
	repeat   = kingpin.Flag("repeat", "repeat").Default("100").Int()
	ncpu     = kingpin.Flag("ncpu", "num of cpu").Default("1").Int()
)

func main() {
	kingpin.Parse()
	ps := parse(*cfgFile)
	cs := create(ps)

	for _, c := range cs {
		writeCfgJSON(c)
		writeCfgIni(c)
		writePbs(c)
	}
	writeCfgs(cs)
	writeQSub(cs)
}

// ParameterSet stores a set of parameters.
type ParameterSet struct {
	Sizes    []int
	NumGenes []int
	LenGenes []int
	Thetas   []float64
	Rhos     []float64
	Deltas   []int
}

func create(par ParameterSet) []simmlst.Config {
	var cfgs []simmlst.Config
	for _, size := range par.Sizes {
		for _, numGene := range par.NumGenes {
			for _, lenGene := range par.LenGenes {
				for _, theta := range par.Thetas {
					for _, rho := range par.Rhos {
						for _, deltas := range par.Deltas {
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

	// add output prefix
	for i := 0; i < len(cfgs); i++ {
		cfgs[i].Output = fmt.Sprintf("%s_individual_%d", *prefix, i)
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

func writeJSON(s interface{}, filename string) {
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	e := json.NewEncoder(w)
	if err := e.Encode(s); err != nil {
		panic(err)
	}
}

func writeCfgJSON(cfg simmlst.Config) {
	filename := cfg.Output + ".cfg.json"
	writeJSON(cfg, filename)
}

func writeCfgIni(cfg simmlst.Config) {
	filename := cfg.Output + ".cfg.ini"
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString(fmt.Sprintf("%s", cfg))
}

func writePbs(c simmlst.Config) {
	filename := c.Output + ".pbs"
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	w.WriteString("#!/bin/bash\n")
	w.WriteString(fmt.Sprintf("#PBS -N %s\n", c.Output))
	w.WriteString(fmt.Sprintf("#PBS -l nodes=1:ppn=%d\n", *ppn))
	w.WriteString(fmt.Sprintf("#PBS -l walltime=%d:00:00\n", *walltime))
	w.WriteString(fmt.Sprintf("#PBS -M ml3365@nyu.edu\n"))
	w.WriteString(fmt.Sprintf("#PBS -m %s\n", *message))
	w.WriteString(fmt.Sprintf("cd %s\n", wd))
	w.WriteString(fmt.Sprintf("%s %s %s --repeat=%d --ncpu=%d\n", *exec, c.Output+".cfg.json", c.Output+".cov.csv", *repeat, *ncpu))
}

func writeQSub(cfgs []simmlst.Config) {
	filename := *prefix + "_qsub.sh"
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("#!/bin/bash\n")
	for _, c := range cfgs {
		w.WriteString(fmt.Sprintf("qsub %s.pbs\n", c.Output))
	}
}

func writeCfgs(cfgs []simmlst.Config) {
	filename := *prefix + "_configs.json"
	writeJSON(cfgs, filename)
}
