package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("no params", func(t *testing.T) {
		err := Copy("", "", 0, 0)
		require.Error(t, err)
	})

	t.Run("invalid file name", func(t *testing.T) {
		err := Copy("noname", "noname", 0, 0)
		require.Error(t, err)
	})

	t.Run("invalid offset", func(t *testing.T) {
		err := Copy("testdata/input.txt", "output.txt", 10000, 0)
		require.EqualError(t, ErrOffsetExceedsFileSize, err.Error())
	})

	t.Run("copy file", func(t *testing.T) {
		err := Copy("testdata/input.txt", "output.txt", 0, 0)

		require.Nil(t, err)

		os.Remove("output.txt")
	})
}
