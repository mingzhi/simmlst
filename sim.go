package simmlst

import (
	"fmt"
	"os/exec"
	"strings"
)

type ParameterSet struct {
	Theta, Rho       float64
	N, Delta         int
	NumGene, LenGene int
	DataFile         string
}

func (p ParameterSet) parse() (options []string) {
	options = append(options, []string{"-N", parseInt(p.N)}...)
	options = append(options, []string{"-D", parseInt(p.Delta)}...)
	options = append(options, []string{"-T", parseFloat64(p.Theta)}...)
	options = append(options, []string{"-R", parseFloat64(p.Rho)}...)
	options = append(options, []string{"-o", p.DataFile}...)

	var blocks []string
	for i := 0; i < p.NumGene; i++ {
		blocks = append(blocks, parseInt(p.LenGene))
	}
	options = append(options, []string{"-B", strings.Join(blocks, ",")}...)

	return
}

func Exec(ps ParameterSet) {
	var options []string
	options = ps.parse()
	cmd := exec.Command("simmlst", options...)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func parseInt(d int) string {
	return fmt.Sprintf("%d", d)
}

func parseFloat64(f float64) string {
	return fmt.Sprintf("%f", f)
}
