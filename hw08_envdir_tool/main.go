package main

import (
	"os"
)

func main() {
	if len(os.Args) < 3 {
		println("Usage: main.exe <path to environment variables> <command> [args]")
		return
	}

	// read environment variables
	env, _ := ReadDir(os.Args[1])

	// run command
	os.Exit(RunCmd(os.Args[2:], env))
}
