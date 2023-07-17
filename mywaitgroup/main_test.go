package main

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func Test_printSomething(t *testing.T) {
	// Save the original stdout to restore it later
	// os.Stdout is a pointer to an os.File struct that represents the standard output globally,
	// which can be used by any part of your program.
	original := os.Stdout

	// Create a pipe to capture the output
	r, w, _ := os.Pipe()

	// Set the standard output to the pipe to capture the output
	// bacause the printSomething function writes to the standard output (i.e., prints to the terminal).
	os.Stdout = w

	var wg sync.WaitGroup
	wg.Add(1)

	want := "epsilon"
	go printSomething(want, &wg)
	wg.Wait()

	// Close the pipe to release the captured output
	// because subsequent read operation can be blocked until the write operation is done.
	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	// Restore the original stdout
	// because other parts of your program (or other tests) might also want to write to the standard output,
	// and they expect it to behave in the standard way (i.e., print to the terminal), not be captured by a pipe.
	os.Stdout = original

	if !strings.Contains(output, want) {
		t.Errorf("Expected to find %s, in %s but not found", want, output)
	}
}
