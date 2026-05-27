package model

import "testing"

func TestVersionString_NoCommit(t *testing.T) {
	// Save and restore original value.
	orig := GitCommit
	defer func() { GitCommit = orig }()

	GitCommit = ""
	if got := VersionString(); got != Version {
		t.Errorf("VersionString() = %q, want %q (no commit)", got, Version)
	}
}

func TestVersionString_WithCommit(t *testing.T) {
	orig := GitCommit
	defer func() { GitCommit = orig }()

	GitCommit = "abc1234"
	want := Version + "+" + GitCommit
	if got := VersionString(); got != want {
		t.Errorf("VersionString() = %q, want %q", got, want)
	}
}
