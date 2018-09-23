package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/IgaguriMK/scanJson/ioutil"
	"github.com/IgaguriMK/scanJson/scanner"
)

func main() {
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

	dec := json.NewDecoder(f)

	value, err := scanner.ParseValue(dec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed decode: ", err)
		os.Exit(1)
	}

	value.Print(os.Stdout)
}
