package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	env := Environment{
		"BAR":   EnvValue{Value: "bar", NeedRemove: false},
		"EMPTY": EnvValue{Value: "", NeedRemove: false},
		"FOO":   EnvValue{Value: "   foo\nwith new line", NeedRemove: false},
		"HELLO": EnvValue{Value: "\"hello\"", NeedRemove: false},
		"UNSET": EnvValue{Value: "", NeedRemove: true},
	}

	t.Run("simple command without params", func(t *testing.T) {
		code := RunCmd([]string{"ping"}, env)
		require.Equal(t, 1, code)
	})

	t.Run("simple command without env", func(t *testing.T) {
		code := RunCmd([]string{"ping"}, nil)
		require.Equal(t, 1, code)
	})

	t.Run("simple command", func(t *testing.T) {
		code := RunCmd([]string{"ping", "127.0.0.1"}, env)
		require.Equal(t, 0, code)
	})
}
