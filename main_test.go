package main

import "testing"

const testInputFile = "test.pb.go"

func TestParseWriteFile(t *testing.T) {
	extracted, err := parseFile(testInputFile)
	if err != nil {
		t.Fatal(err)
	}

	if err = writeFile(testInputFile, extracted); err != nil {
		t.Fatal(err)
	}
}
