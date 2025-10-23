package main

import (
	"fmt"
	"os"
)

func main() {
	dirPath := os.Args[1]
	cmd := os.Args[2:]

	envVariables, err := ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading variables in %s failed with error: %v", dirPath, err)
		os.Exit(1)
	}
	os.Exit(RunCmd(cmd, envVariables))
}
