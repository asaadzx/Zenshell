package main

import (
	"fmt"
	"os"

	"bakshell/internal/shell"
)

func main() {
	s, err := shell.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(s.Run())
}
