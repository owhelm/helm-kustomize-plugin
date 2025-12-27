package main

import (
	"io"
	"os"
)

func main() {
	// For now, just pass through stdin to stdout
	// This validates the plugin structure works
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		os.Stderr.WriteString("Error reading input: " + err.Error() + "\n")
		os.Exit(1)
	}
}