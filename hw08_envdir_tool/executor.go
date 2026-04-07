package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	ErrCodeWrongParams  = 1
	ErrCodeUnknownError = 127
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return ErrCodeWrongParams
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...) //nolint:gosec

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
	var ee *exec.ExitError
	err := command.Run()
	if err != nil {
		fmt.Println(err)

		if errors.As(err, &ee) {
			return ee.ExitCode()
		}

		return ErrCodeUnknownError
	}

	return 0
}

// Convert key=value string array to map.
func KeyValueArrayToMap(kv []string) map[string]string {
	env := make(map[string]string)
	for _, kv := range kv {
		s := strings.Split(kv, "=")
		env[s[0]] = s[1]
	}
	return env
}
