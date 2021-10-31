package main

import (
	"os"
	"testing"
)

func TestGetBytesFromFile(t *testing.T) {
	fileName := "testing_get_bytes"
	fileContent := "The quick brown fox jumps over the lazy dog"
	file, err := os.Create(fileName)
	defer os.Remove(fileName)
	assertNil(t, err)
	_, err = file.WriteString(fileContent)
	assertNil(t, err)
	file.Close()

	bytes, err := getBytesFromFile(fileName)
	assertNil(t, err)
	assertEqual(t, fileContent, string(bytes))
}

func TestGetBytesFromFileError(t *testing.T) {
	fileName := "testing_get_bytes"
	_, err := getBytesFromFile(fileName)
	assertNotNil(t, err)
}

func TestComputeHash(t *testing.T) {
	bytes := []byte("The quick brown fox jumps over the lazy dog")
	expect := "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"
	result := computeHash(bytes)
	assertEqual(t, expect, result)
}

func assertNil(t *testing.T, value interface{}) {
	if value != nil {
		t.Fatal("Value should be nil\n")
	}
}

func assertNotNil(t *testing.T, value interface{}) {
	if value == nil {
		t.Fatal("Value should not be nil\n")
	}
}

func assertEqual(t *testing.T, expect, result interface{}) {
	if expect != result {
		t.Fatalf("Values are different\n\texpect \t%v\n\tgot\t%v", expect, result)
	}
}
