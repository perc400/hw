package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedFilename = errors.New("unsupported filename (contains \"=\" symbol)")
	ErrDirectoryPath       = errors.New("not a directory")
	ErrNotAFile            = errors.New("is a directory")
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	environmentVariables := make(Environment)

	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(dir, fileName)

		if strings.Contains(fileName, "=") {
			return nil, fmt.Errorf("file %s [%w]", fileName, ErrUnsupportedFilename)
		}

		fileInfo, err := os.Lstat(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %s info [%w]", fileName, err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("%s is a directory [%w]", fileName, ErrNotAFile)
		}

		if fileInfo.Size() == 0 {
			environmentVariables[fileName] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}
			continue
		}

		fileContent, err := os.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			return nil, fmt.Errorf("reading file %s [%w]", fileName, err)
		}

		fileLines := strings.Split(string(fileContent), "\n")
		variableValue := strings.TrimRight(fileLines[0], " \t")

		variableValue = string(bytes.ReplaceAll([]byte(variableValue), []byte("\x00"), []byte("\n")))

		environmentVariables[fileName] = EnvValue{
			Value:      variableValue,
			NeedRemove: false,
		}
	}

	return environmentVariables, nil
}
