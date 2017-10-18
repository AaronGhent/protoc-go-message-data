package main

import (
	"flag"
	"log"
)

func main() {
	var inputFile string

	flag.StringVar(&inputFile, "input", "", "path to input file")

	flag.Parse()

	if len(inputFile) == 0 {
		log.Fatal("input file is mandatory")
	}

	extracted, err := parseFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	if err = writeFile(inputFile, extracted); err != nil {
		log.Fatal(err)
	}
}
