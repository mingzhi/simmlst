package io

import (
	"bufio"
	"bytes"
	"github.com/mingzhi/biogo/seq"
	"io"
	"os"
)

func ReadXMFA(filename string) [][]*seq.Sequence {
	seqGroups := [][]*seq.Sequence{}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var b bytes.Buffer

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		if line[0] != '=' {
			b.Write(line)
		} else {
			sequences := readFasta(&b)
			seqGroups = append(seqGroups, sequences)
			b = *new(bytes.Buffer)
		}
	}

	return seqGroups
}

func readFasta(b io.Reader) []*seq.Sequence {
	fastaReader := seq.NewFastaReader(b)
	sequences, err := fastaReader.ReadAll()
	if err != nil {
		panic(err)
	}

	return sequences
}
