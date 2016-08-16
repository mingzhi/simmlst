package simmlst

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Config stores a set of population parameters.
type Config struct {
	Theta, Rho       float64
	N, Delta         int
	NumGene, LenGene int
	Output           string
}

func (p Config) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "theta = %g\n", p.Theta)
	fmt.Fprintf(&b, "rho = %g\n", p.Rho)
	fmt.Fprintf(&b, "n = %d\n", p.N)
	fmt.Fprintf(&b, "num_gene = %d\n", p.NumGene)
	fmt.Fprintf(&b, "len_gene = %d\n", p.LenGene)
	fmt.Fprintf(&b, "output = %s\n", p.Output)

	return b.String()
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

// Exec run simmlst.
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
