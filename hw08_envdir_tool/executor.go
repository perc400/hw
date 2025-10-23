package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func prepareEnv(cmd *exec.Cmd, env Environment) {
	cmd.Env = os.Environ()

	processedMap := make(map[string]string)
	order := make([]string, 0, len(cmd.Env))

	for _, keyValue := range cmd.Env {
		pair := strings.SplitN(keyValue, "=", 2)
		if len(pair) < 2 {
			continue
		}
		k, v := pair[0], pair[1]
		processedMap[k] = v
		order = append(order, k)
	}

	for k, v := range env {
		if v.NeedRemove {
			delete(processedMap, k)
		} else {
			if _, exists := processedMap[k]; !exists {
				order = append(order, k)
			}
			processedMap[k] = v.Value
		}
	}

	newEnv := make([]string, 0, cap(cmd.Env))
	for _, key := range order {
		value, ok := processedMap[key]
		if ok {
			newEnv = append(newEnv, key+"="+value)
		}
	}

	cmd.Env = newEnv
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	binaryPath, err := exec.LookPath(cmd[0])
	if err != nil {
		return 127
	}
	command := exec.Command(binaryPath, cmd[1:]...)

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	prepareEnv(command, env)

	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}
