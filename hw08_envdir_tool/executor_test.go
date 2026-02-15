package main

import "testing"

func TestRunCmd(t *testing.T) {
	env := Environment{
		"BAR":   EnvValue{Value: "bar", NeedRemove: false},
		"EMPTY": EnvValue{Value: "", NeedRemove: false},
		"FOO":   EnvValue{Value: "   foo\nwith new line", NeedRemove: false},
		"HELLO": EnvValue{Value: "\"hello\"", NeedRemove: false},
		"UNSET": EnvValue{Value: "", NeedRemove: true},
	}

	t.Run("simple command without params", func(t *testing.T) {
		RunCmd([]string{"ping"}, env)
	})

	t.Run("simple command", func(t *testing.T) {
		RunCmd([]string{"ping", "127.0.0.1"}, env)
	})
}
