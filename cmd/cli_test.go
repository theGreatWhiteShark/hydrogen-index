package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandDefaultsToScan(t *testing.T) {
	root := NewRootCmd()
	b := new(bytes.Buffer)
	root.SetOut(b)
	root.SetArgs([]string{"--dir", "../res/hydrogen-artifacts/legacy-patterns", "--out", "test-root.json"})

	err := root.Execute()
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Scanning directory: ../res/hydrogen-artifacts/legacy-patterns")
	os.Remove("test-root.json")
}

func TestScanFlags(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
		out      string
	}{
		{[]string{"scan", "--dir", "../res/hydrogen-artifacts/legacy-patterns", "--out", "test1.json"}, "../res/hydrogen-artifacts/legacy-patterns", "test1.json"},
		{[]string{"scan", "-d", "../res/hydrogen-artifacts/legacy-songs", "-o", "test2.json"}, "../res/hydrogen-artifacts/legacy-songs", "test2.json"},
	}

	for _, c := range cases {
		root := NewRootCmd()
		b := new(bytes.Buffer)
		root.SetOut(b)
		root.SetArgs(c.args)
		
		err := root.Execute()
		assert.NoError(t, err)

		assert.Contains(t, b.String(), "Scanning directory: "+c.expected)
		os.Remove(c.out)
	}
}

func TestValidateArgs(t *testing.T) {
	// Create a dummy file for validation
	dummyFile := "dummy.json"
	os.WriteFile(dummyFile, []byte(`{"version":"0.1.0","created":"","patternCount":0,"songCount":0,"drumkitCount":0,"patterns":[],"songs":[],"drumkits":[],"hash":""}`), 0644)
	defer os.Remove(dummyFile)

	cases := []struct {
		args []string
		err  bool
	}{
		{[]string{"validate", dummyFile}, false},
		{[]string{"validate"}, true},
		{[]string{"validate", "file1.json", "file2.json"}, true},
	}

	for _, c := range cases {
		root := NewRootCmd()
		b := new(bytes.Buffer)
		root.SetOut(b)
		root.SetArgs(c.args)

		err := root.Execute()
		if c.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Contains(t, b.String(), "Validating "+dummyFile)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	root := NewRootCmd()
	b := new(bytes.Buffer)
	root.SetOut(b)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "hydrogen-index v0.1.0")
}
