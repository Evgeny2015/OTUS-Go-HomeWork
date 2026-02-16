package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
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
	dirEntry, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// create map
	env := make(Environment)

	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue
		}

		// open file
		file, err := os.Open(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		defer file.Close()

		reader := bufio.NewReader(file)

		// read the first line
		line, _ := reader.ReadString('\n')
		if line == "" {
			// append to map with NeedRemove = true
			env[entry.Name()] = EnvValue{Value: "", NeedRemove: true}
			continue
		}

		// trim right the line
		line = strings.TrimRight(line, "\n ")

		// replace \x00 with \n
		line = strings.ReplaceAll(line, "\x00", "\n")

		// append to map with NeedRemove = false
		env[entry.Name()] = EnvValue{Value: line, NeedRemove: false}
	}

	return env, nil
}
