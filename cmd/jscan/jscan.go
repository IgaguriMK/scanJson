package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/IgaguriMK/scanJson/scanner"
)

func main() {
	var useGzip bool
	flag.BoolVar(&useGzip, "d", false, "Decompress gzip.")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: no arguments")
		os.Exit(1)
	}

	filename := args[0]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error can't open file: ", err)
		os.Exit(1)
	}
	defer f.Close()

	var r io.Reader = f

	if useGzip {
		gr, err := gzip.NewReader(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Can't open as gzipped file: ", err)
			os.Exit(1)
		}
		defer gr.Close()
		r = gr
	}

	dec := json.NewDecoder(r)

	value, err := scanner.ParseValue(dec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed decode: ", err)
		os.Exit(1)
	}

	value.Print(os.Stdout)
}
