package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {

	command := exec.Command(cmd[0], cmd[1:]...)

	// get system environment variables
	envMap := KeyValueArrayToMap(os.Environ())

	// append/delete environment variables to command
	for k, v := range env {
		if v.NeedRemove {
			delete(envMap, k)
		} else {
			envMap[k] = v.Value
		}
	}

	// convert map to array
	cmdEnv := []string{}
	for k, v := range envMap {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	command.Env = cmdEnv

	// set stdin, stdout, stderr
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	// run command
	err := command.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		} else {
			return 127
		}
	}

	return 0
}

// convert key=value string array to map
func KeyValueArrayToMap(kv []string) map[string]string {
	env := make(map[string]string)
	for _, kv := range kv {
		s := strings.Split(kv, "=")
		env[s[0]] = s[1]
	}
	return env
}
