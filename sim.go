package simmlst

import (
	"fmt"
	"os/exec"
	"strings"
)

type Config struct {
	Theta, Rho       float64
	N, Delta         int
	NumGene, LenGene int
}

func (p Config) parse() (options []string) {
	options = append(options, []string{"-N", parseInt(p.N)}...)
	options = append(options, []string{"-D", parseInt(p.Delta)}...)
	options = append(options, []string{"-T", parseFloat64(p.Theta)}...)
	options = append(options, []string{"-R", parseFloat64(p.Rho)}...)

	var blocks []string
	for i := 0; i < p.NumGene; i++ {
		blocks = append(blocks, parseInt(p.LenGene))
	}
	options = append(options, []string{"-B", strings.Join(blocks, ",")}...)

	return
}

func Exec(ps Config, tempfile string) {
	var options []string
	options = ps.parse()
	options = append(options, []string{"-o", tempfile}...)
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
