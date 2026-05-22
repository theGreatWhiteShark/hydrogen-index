package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/index"
	"github.com/hydrogen-music/hydrogen-index/internal/model"
	"github.com/hydrogen-music/hydrogen-index/internal/scanner"
	"github.com/hydrogen-music/hydrogen-index/internal/validator"
)

// repoRoot returns the absolute path to the repository root, derived from the
// location of this source file rather than the working directory, so the path
// is correct regardless of where `go test` is invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// This file lives at <root>/cmd/integration_test.go; parent of cmd/ is root.
	return filepath.Dir(filepath.Dir(file))
}

// TestIntegration_ScanAndValidate exercises the full pipeline:
// scan → build → finalize → write → validate → assert structural correctness.
func TestIntegration_ScanAndValidate(t *testing.T) {
	artifactsDir := filepath.Join(repoRoot(t), "res", "hydrogen-artifacts")

	// ── Scan ─────────────────────────────────────────────────────────────────
	artifacts, errs := scanner.Scan(artifactsDir, "")
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// ── Build ─────────────────────────────────────────────────────────────────
	idx, err := index.Build(artifacts)
	if err != nil {
		t.Fatalf("index.Build: %v", err)
	}

	// ── Finalize ──────────────────────────────────────────────────────────────
	data, err := index.Finalize(idx)
	if err != nil {
		t.Fatalf("index.Finalize: %v", err)
	}

	// ── Write to temp file ────────────────────────────────────────────────────
	f, err := os.CreateTemp("", "index-integration-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpPath := f.Name()
	t.Cleanup(func() { os.Remove(tmpPath) })

	if _, err := f.Write(data); err != nil {
		f.Close()
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	// ── Validate against schema ───────────────────────────────────────────────
	if err := validator.Validate(tmpPath); err != nil {
		t.Fatalf("schema validation failed: %v", err)
	}

	// ── Unmarshal and assert structural correctness ───────────────────────────
	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Top-level count fields must match the expected fixture totals.
	if result.PatternCount != 3 {
		t.Errorf("PatternCount = %d, want 3", result.PatternCount)
	}
	if result.SongCount != 16 {
		t.Errorf("SongCount = %d, want 16", result.SongCount)
	}
	if result.DrumkitCount != 2 {
		t.Errorf("DrumkitCount = %d, want 2", result.DrumkitCount)
	}

	// Slice lengths must agree with the declared counts.
	if len(result.Patterns) != result.PatternCount {
		t.Errorf("len(Patterns) = %d, want %d", len(result.Patterns), result.PatternCount)
	}
	if len(result.Songs) != result.SongCount {
		t.Errorf("len(Songs) = %d, want %d", len(result.Songs), result.SongCount)
	}
	if len(result.Drumkits) != result.DrumkitCount {
		t.Errorf("len(Drumkits) = %d, want %d", len(result.Drumkits), result.DrumkitCount)
	}

	// SHA-256 hex digest is always exactly 64 characters.
	if len(result.Hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(result.Hash))
	}

	if result.Version != model.Version {
		t.Errorf("Version = %q, want %q", result.Version, model.Version)
	}
	if result.Created == "" {
		t.Error("Created must not be empty")
	}

	// ── Per-entry field assertions ────────────────────────────────────────────
	for i, p := range result.Patterns {
		if p.Type != model.ArtifactTypePattern {
			t.Errorf("Patterns[%d].Type = %q, want %q", i, p.Type, model.ArtifactTypePattern)
		}
		if p.Name == "" {
			t.Errorf("Patterns[%d].Name is empty", i)
		}
		if p.URL == "" {
			t.Errorf("Patterns[%d].URL is empty", i)
		}
		if len(p.Hash) != 64 {
			t.Errorf("Patterns[%d].Hash length = %d, want 64", i, len(p.Hash))
		}
		if p.Size <= 0 {
			t.Errorf("Patterns[%d].Size = %d, want > 0", i, p.Size)
		}
		if p.Notes < 0 {
			t.Errorf("Patterns[%d].Notes = %d, want >= 0", i, p.Notes)
		}
		if p.Tags == nil {
			t.Errorf("Patterns[%d].Tags must not be nil", i)
		}
		if p.InstrumentTypes == nil {
			t.Errorf("Patterns[%d].InstrumentTypes must not be nil", i)
		}
	}

	for i, s := range result.Songs {
		if s.Type != model.ArtifactTypeSong {
			t.Errorf("Songs[%d].Type = %q, want %q", i, s.Type, model.ArtifactTypeSong)
		}
		if s.Name == "" {
			t.Errorf("Songs[%d].Name is empty", i)
		}
		if s.Patterns < 0 {
			t.Errorf("Songs[%d].Patterns = %d, want >= 0", i, s.Patterns)
		}
	}

	for i, d := range result.Drumkits {
		if d.Type != model.ArtifactTypeDrumkit {
			t.Errorf("Drumkits[%d].Type = %q, want %q", i, d.Type, model.ArtifactTypeDrumkit)
		}
		if d.Instruments <= 0 {
			t.Errorf("Drumkits[%d].Instruments = %d, want > 0", i, d.Instruments)
		}
		if d.Samples < 0 {
			t.Errorf("Drumkits[%d].Samples = %d, want >= 0", i, d.Samples)
		}
		if d.Components < 0 {
			t.Errorf("Drumkits[%d].Components = %d, want >= 0", i, d.Components)
		}
		if d.InstrumentTypes == nil {
			t.Errorf("Drumkits[%d].InstrumentTypes must not be nil", i)
		}
	}

	// ── Spot-checks for known fixture values ──────────────────────────────────

	// The v2.0.0 pattern fixture has a known set of instrument types and note count.
	var v200Pattern *model.PatternEntry
	for i := range result.Patterns {
		if result.Patterns[i].URL == "v2.0.0.h2pattern" {
			v200Pattern = &result.Patterns[i]
			break
		}
	}
	if v200Pattern == nil {
		t.Error("v2.0.0.h2pattern not found in index")
	} else {
		wantTypes := []string{"Hand Clap", "Kick", "Snare", "Stick"}
		if len(v200Pattern.InstrumentTypes) != len(wantTypes) {
			t.Errorf("v2.0.0 pattern InstrumentTypes = %v, want %v",
				v200Pattern.InstrumentTypes, wantTypes)
		} else {
			for i, typ := range wantTypes {
				if v200Pattern.InstrumentTypes[i] != typ {
					t.Errorf("v2.0.0 pattern InstrumentTypes[%d] = %q, want %q",
						i, v200Pattern.InstrumentTypes[i], typ)
				}
			}
		}
		if v200Pattern.Notes != 20 {
			t.Errorf("v2.0.0 pattern Notes = %d, want 20", v200Pattern.Notes)
		}
	}

	// The v2.0.0 drumkit (testKit, from tar) has known instrument/sample counts and tags.
	var v200Drumkit *model.DrumkitEntry
	for i := range result.Drumkits {
		if result.Drumkits[i].URL == "v2.0.0.h2drumkit" {
			v200Drumkit = &result.Drumkits[i]
			break
		}
	}
	if v200Drumkit == nil {
		t.Error("v2.0.0.h2drumkit not found in index")
	} else {
		if v200Drumkit.Instruments != 3 {
			t.Errorf("v2.0.0 drumkit Instruments = %d, want 3", v200Drumkit.Instruments)
		}
		if v200Drumkit.Samples != 3 {
			t.Errorf("v2.0.0 drumkit Samples = %d, want 3", v200Drumkit.Samples)
		}
		wantTags := []string{"Example", "Drumkit"}
		if len(v200Drumkit.Tags) != len(wantTags) {
			t.Errorf("v2.0.0 drumkit Tags = %v, want %v", v200Drumkit.Tags, wantTags)
		} else {
			for i, tag := range wantTags {
				if v200Drumkit.Tags[i] != tag {
					t.Errorf("v2.0.0 drumkit Tags[%d] = %q, want %q",
						i, v200Drumkit.Tags[i], tag)
				}
			}
		}
	}
}

// TestIntegration_ValidateReferenceFile confirms that the canonical reference
// file in the repository passes schema validation.
func TestIntegration_ValidateReferenceFile(t *testing.T) {
	refPath := filepath.Join(repoRoot(t), "res", "references-index.json")
	if err := validator.Validate(refPath); err != nil {
		t.Fatalf("references-index.json failed validation: %v", err)
	}
}

// TestIntegration_DefaultScan builds the binary and runs it with no flags,
// verifying that git-root auto-detection and default output path both work.
func TestIntegration_DefaultScan(t *testing.T) {
	root := repoRoot(t)

	// Build the binary into a temp directory so it is cleaned up automatically.
	tmpDir, err := os.MkdirTemp("", "hydrogen-index-bin-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	binPath := filepath.Join(tmpDir, "hydrogen-index")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	// The generated index.json lands in the git root by default.
	// Register cleanup before running so a partial output is still removed on failure.
	outputPath := filepath.Join(root, "index.json")
	t.Cleanup(func() { os.Remove(outputPath) })

	// Run the binary from the repo root so findGitRoot() locates .git immediately.
	run := exec.Command(binPath)
	run.Dir = root
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created, but it does not exist", outputPath)
	}

	// The generated file must also pass schema validation.
	if err := validator.Validate(outputPath); err != nil {
		t.Errorf("generated index.json failed validation: %v", err)
	}
}

// TestIntegration_ProviderGitHub verifies that --provider github constructs
// correct GitHub raw-file URLs.
func TestIntegration_ProviderGitHub(t *testing.T) {
	root := repoRoot(t)

	tmpDir, err := os.MkdirTemp("", "hydrogen-index-bin-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	binPath := filepath.Join(tmpDir, "hydrogen-index")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	outputPath := filepath.Join(tmpDir, "index.json")
	artifactsDir := filepath.Join(root, "res", "hydrogen-artifacts")
	run := exec.Command(binPath, "-d", artifactsDir, "-o", outputPath,
		"--provider", "github", "--repo", "user/repo", "--branch", "main")
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// All URLs should start with the GitHub raw base URL.
	prefix := "https://raw.githubusercontent.com/user/repo/main/"
	for i, p := range result.Patterns {
		if !strings.HasPrefix(p.URL, prefix) {
			t.Errorf("Patterns[%d].URL = %q, want prefix %q", i, p.URL, prefix)
		}
	}
	for i, s := range result.Songs {
		if !strings.HasPrefix(s.URL, prefix) {
			t.Errorf("Songs[%d].URL = %q, want prefix %q", i, s.URL, prefix)
		}
	}
	for i, d := range result.Drumkits {
		if !strings.HasPrefix(d.URL, prefix) {
			t.Errorf("Drumkits[%d].URL = %q, want prefix %q", i, d.URL, prefix)
		}
	}
}

// TestIntegration_ProviderGitLab verifies that --provider gitlab constructs
// correct GitLab raw-file URLs.
func TestIntegration_ProviderGitLab(t *testing.T) {
	root := repoRoot(t)

	tmpDir, err := os.MkdirTemp("", "hydrogen-index-bin-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	binPath := filepath.Join(tmpDir, "hydrogen-index")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	outputPath := filepath.Join(tmpDir, "index.json")
	artifactsDir := filepath.Join(root, "res", "hydrogen-artifacts")
	run := exec.Command(binPath, "-d", artifactsDir, "-o", outputPath,
		"--provider", "gitlab", "--repo", "user/repo", "--branch", "develop")
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// All URLs should start with the GitLab raw base URL.
	prefix := "https://gitlab.com/user/repo/-/raw/develop/"
	for i, p := range result.Patterns {
		if !strings.HasPrefix(p.URL, prefix) {
			t.Errorf("Patterns[%d].URL = %q, want prefix %q", i, p.URL, prefix)
		}
	}
}

// TestIntegration_BaseURL verifies that --base-url prepends arbitrary URLs.
func TestIntegration_BaseURL(t *testing.T) {
	root := repoRoot(t)

	tmpDir, err := os.MkdirTemp("", "hydrogen-index-bin-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	binPath := filepath.Join(tmpDir, "hydrogen-index")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	outputPath := filepath.Join(tmpDir, "index.json")
	artifactsDir := filepath.Join(root, "res", "hydrogen-artifacts")
	run := exec.Command(binPath, "-d", artifactsDir, "-o", outputPath,
		"--base-url", "https://example.com/artifacts")
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	prefix := "https://example.com/artifacts/"
	for i, p := range result.Patterns {
		if !strings.HasPrefix(p.URL, prefix) {
			t.Errorf("Patterns[%d].URL = %q, want prefix %q", i, p.URL, prefix)
		}
	}
}

// TestIntegration_NoBaseURL verifies that without URL flags, URLs are relative paths.
func TestIntegration_NoBaseURL(t *testing.T) {
	root := repoRoot(t)

	tmpDir, err := os.MkdirTemp("", "hydrogen-index-bin-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	binPath := filepath.Join(tmpDir, "hydrogen-index")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	outputPath := filepath.Join(tmpDir, "index.json")
	artifactsDir := filepath.Join(root, "res", "hydrogen-artifacts")
	run := exec.Command(binPath, "-d", artifactsDir, "-o", outputPath)
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// URLs should be plain relative paths (no "http" prefix).
	for i, p := range result.Patterns {
		if strings.HasPrefix(p.URL, "http") {
			t.Errorf("Patterns[%d].URL = %q, want relative path (no http prefix)", i, p.URL)
		}
	}
}
