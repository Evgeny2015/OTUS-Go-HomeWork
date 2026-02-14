package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {

	result := Environment{
		"BAR":   EnvValue{Value: "bar", NeedRemove: false},
		"EMPTY": EnvValue{Value: "", NeedRemove: false},
		"FOO":   EnvValue{Value: "   foo\nwith new line", NeedRemove: false},
		"HELLO": EnvValue{Value: "\"hello\"", NeedRemove: false},
		"UNSET": EnvValue{Value: "", NeedRemove: true},
	}

	t.Run("test invalid folder name", func(t *testing.T) {
		_, err := ReadDir("invalid")

		require.Error(t, err)
	})

	t.Run("test get env", func(t *testing.T) {
		env, err := ReadDir("testdata/env")

		require.NoError(t, err)
		require.Equal(t, result, env)
	})
}
