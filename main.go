package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/enolgor/muxc/muxc"
)

var file string

func init() {
	flag.StringVar(&file, "f", "muxc.yaml", "path to yaml configuration file")
	flag.Parse()
}

func main() {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate %s file: %s\n", file, err.Error())
		os.Exit(1)
	}
	defer f.Close()
	dir := filepath.Dir(file)
	if err = muxc.Generate(file, f, dir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating muxc routes: %s\n", err.Error())
		os.Exit(1)
	}
}
