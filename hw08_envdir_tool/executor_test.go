package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("no_command_provided", func(t *testing.T) {
		envVariables, err := ReadDir(testdataDir)
		require.NoError(t, err)

		cmd := []string{}
		returnCode := RunCmd(cmd, envVariables)
		require.Equal(t, returnCode, 1)
	})

	t.Run("simple_successful_command", func(t *testing.T) {
		env, err := ReadDir(testdataDir)
		require.NoError(t, err)

		cmd := []string{"true"}
		returnCode := RunCmd(cmd, env)
		require.Equal(t, 0, returnCode)
	})

	t.Run("nonexistent_command", func(t *testing.T) {
		env, err := ReadDir(testdataDir)
		require.NoError(t, err)

		cmd := []string{"gfsdlgfsd"}
		returnCode := RunCmd(cmd, env)
		require.Equal(t, 127, returnCode)
	})
}
