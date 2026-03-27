package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
	internalvalidate "github.com/theGreatWhiteShark/hydrogen-index/internal/validate"
)

func TestRootCommandDefaultsToScan(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2pattern",
		filepath.Join(repoRoot, "artifacts", "v2.0.0.h2pattern"),
	)

	workingDir := filepath.Join(repoRoot, "nested", "working")
	mustMkdirAll(t, workingDir)

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	command := NewRootCommand(Dependencies{
		WorkingDir: workingDir,
		Stdout:     stdout,
		Stderr:     stderr,
		Version:    "0.1.0",
		Now:        func() time.Time { return time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC) },
	})
	command.SetArgs([]string{"--base-url", "https://example.com/content"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	outputPath := filepath.Join(workingDir, "index.json")
	index := readIndexFile(t, outputPath)

	if index.PatternCount != 1 || len(index.Patterns) != 1 {
		t.Fatalf("unexpected pattern counts: %+v", index)
	}

	if got, want := index.Patterns[0].URL, "https://example.com/content/artifacts/v2.0.0.h2pattern"; got != want {
		t.Fatalf("unexpected artifact URL: got %q want %q", got, want)
	}

	if err := internalvalidate.ValidateBytes(mustReadFile(t, outputPath)); err != nil {
		t.Fatalf("generated index did not validate: %v", err)
	}

	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	}

func TestScanDirectoryFlagSkipsGitLookupAndRespectsRelativeOutputPath(t *testing.T) {
	t.Parallel()

	scanDir := t.TempDir()
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-patterns/legacy_pattern.h2pattern",
		filepath.Join(scanDir, "legacy_pattern.h2pattern"),
	)

	workingDir := t.TempDir()
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	command := NewRootCommand(Dependencies{
		WorkingDir: workingDir,
		Stdout:     stdout,
		Stderr:     stderr,
		Version:    "0.1.0",
		Now:        func() time.Time { return time.Date(2026, 3, 27, 13, 0, 0, 0, time.UTC) },
	})
	command.SetArgs([]string{
		"scan",
		"-d", scanDir,
		"-o", filepath.Join("custom", "feed.json"),
		"--base-url", "https://example.com/library",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	outputPath := filepath.Join(workingDir, "custom", "feed.json")
	index := readIndexFile(t, outputPath)

	if index.PatternCount != 1 || len(index.Patterns) != 1 {
		t.Fatalf("unexpected pattern counts: %+v", index)
	}

	if got, want := index.Patterns[0].URL, "https://example.com/library/legacy_pattern.h2pattern"; got != want {
		t.Fatalf("unexpected artifact URL: got %q want %q", got, want)
	}

	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestVersionFlagPrintsVersionWithoutRequiringScanFlags(t *testing.T) {
	t.Parallel()

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	command := NewRootCommand(Dependencies{
		WorkingDir: t.TempDir(),
		Stdout:     stdout,
		Stderr:     stderr,
		Version:    "0.1.0",
		Now:        time.Now,
	})
	command.SetArgs([]string{"--version"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "0.1.0") {
		t.Fatalf("version output %q does not contain version", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestHelpFlagShowsCommands(t *testing.T) {
	t.Parallel()

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	command := NewRootCommand(Dependencies{
		WorkingDir: t.TempDir(),
		Stdout:     stdout,
		Stderr:     stderr,
		Version:    "0.1.0",
		Now:        time.Now,
	})
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdout.String()
	for _, fragment := range []string{"scan", "validate", "--base-url"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("help output %q does not contain %q", output, fragment)
		}
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestValidateCommandPrintsSuccessForValidIndex(t *testing.T) {
	t.Parallel()

	workingDir := t.TempDir()
	indexPath := filepath.Join(workingDir, "index.json")
	mustCopyFile(t, "/home/phil/git/hydrogen-index/res/references-index.json", indexPath)

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	command := NewRootCommand(Dependencies{
		WorkingDir: workingDir,
		Stdout:     stdout,
		Stderr:     stderr,
		Version:    "0.1.0",
		Now:        time.Now,
	})
	command.SetArgs([]string{"validate", indexPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "valid" {
		t.Fatalf("unexpected validate output: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func readIndexFile(t *testing.T, path string) indexfile.Document {
	t.Helper()

	data := mustReadFile(t, path)
	var document indexfile.Document
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("json.Unmarshal(%q) failed: %v", path, err)
	}
	return document
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) failed: %v", path, err)
	}
	return data
}

func mustCopyFile(t *testing.T, sourcePath string, destinationPath string) {
	t.Helper()

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) failed: %v", sourcePath, err)
	}

	mustMkdirAll(t, filepath.Dir(destinationPath))
	if err := os.WriteFile(destinationPath, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", destinationPath, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) failed: %v", path, err)
	}
}
