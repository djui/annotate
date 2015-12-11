package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrefix(t *testing.T) {
	assert.Equal(t, "", formatPrefix("", os.Stdout, "foo"))

	assert.Equal(t, "%", formatPrefix("%", os.Stdout, "foo"))
	assert.Equal(t, "%", formatPrefix("%%", os.Stdout, "foo"))
	assert.Equal(t, "%%", formatPrefix("%%%", os.Stdout, "foo"))
	assert.Equal(t, "%%", formatPrefix("%%%%", os.Stdout, "foo"))

	assert.Equal(t, "O", formatPrefix("%>", os.Stdout, "foo"))
	assert.Equal(t, "E", formatPrefix("%>", os.Stderr, "foo"))
	assert.Equal(t, "I", formatPrefix("%>", os.Stdin, "foo"))
	assert.Equal(t, "?", formatPrefix("%>", os.NewFile(3, "fd3"), "foo"))

	assert.Equal(t, "foo", formatPrefix("%0", os.Stdout, "foo"))
	assert.Equal(t, "foo", formatPrefix("%0", os.Stdout, "foo"))

	// s := formatPrefix("%Y-%m-%dT%H:%M:%S.%N", os.Stdout, "foo")
	// _, err := time.Parse(time.RFC3339Nano, s)
	// assert.NoError(t, err)
	// s = formatPrefix("%FT%T", os.Stdout, "foo")
	// _, err = time.Parse(time.RFC3339, s)
	// assert.NoError(t, err)
}
