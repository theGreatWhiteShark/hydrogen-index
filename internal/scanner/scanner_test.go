package scanner_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
	"github.com/hydrogen-music/hydrogen-index/internal/scanner"
)

// artifactsDir returns the absolute path to the shared test-fixture directory.
// Using runtime.Caller keeps the path correct regardless of where `go test` is
// invoked from.
func artifactsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "res", "hydrogen-artifacts")
}

func TestScan(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, errs := scanner.Scan(dir)

	// Parse errors are non-fatal but unexpected for well-formed fixtures.
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	var patterns, drumkits, songs []scanner.ArtifactFile
	for _, a := range artifacts {
		switch a.Metadata.(type) {
		case *model.PatternMetadata:
			patterns = append(patterns, a)
		case *model.DrumkitMetadata:
			drumkits = append(drumkits, a)
		case *model.SongMetadata:
			songs = append(songs, a)
		default:
			t.Errorf("unexpected metadata type %T for %s", a.Metadata, a.RelPath)
		}
	}

	if got, want := len(patterns), 3; got != want {
		t.Errorf("pattern count: got %d, want %d", got, want)
	}
	if got, want := len(drumkits), 4; got != want {
		t.Errorf("drumkit count: got %d, want %d", got, want)
	}
	if got, want := len(songs), 16; got != want {
		t.Errorf("song count: got %d, want %d", got, want)
	}

	for _, a := range artifacts {
		// SHA-256 hex digest is always 64 characters.
		if got := len(a.Hash); got != 64 {
			t.Errorf("%s: hash length %d, want 64", a.RelPath, got)
		}
		if a.Size <= 0 {
			t.Errorf("%s: non-positive size %d", a.RelPath, a.Size)
		}
		if a.RelPath == "" {
			t.Errorf("%s: empty RelPath", a.Path)
		}
		// RelPath must use forward slashes (URL-style).
		for _, ch := range a.RelPath {
			if ch == '\\' {
				t.Errorf("%s: RelPath contains backslash", a.RelPath)
				break
			}
		}
	}
}

// TestScanPatternMetadataTypes verifies the modern v2.0.0 pattern is parsed
// into *model.PatternMetadata with the expected type assertion.
func TestScanPatternMetadataTypes(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir)

	for _, a := range artifacts {
		if a.RelPath == "v2.0.0.h2pattern" {
			if _, ok := a.Metadata.(*model.PatternMetadata); !ok {
				t.Errorf("v2.0.0.h2pattern: got %T, want *model.PatternMetadata", a.Metadata)
			}
			return
		}
	}
	t.Error("v2.0.0.h2pattern not found in scan results")
}

// TestScanDrumkitTarMetadata verifies that a .h2drumkit tar archive is scanned
// correctly and its size/hash reflect the archive file, not the embedded XML.
func TestScanDrumkitTarMetadata(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir)

	for _, a := range artifacts {
		if a.RelPath == "v2.0.0.h2drumkit" {
			if _, ok := a.Metadata.(*model.DrumkitMetadata); !ok {
				t.Errorf("v2.0.0.h2drumkit: got %T, want *model.DrumkitMetadata", a.Metadata)
			}
			if a.Size <= 0 {
				t.Errorf("v2.0.0.h2drumkit: non-positive size %d", a.Size)
			}
			if len(a.Hash) != 64 {
				t.Errorf("v2.0.0.h2drumkit: hash length %d, want 64", len(a.Hash))
			}
			return
		}
	}
	t.Error("v2.0.0.h2drumkit not found in scan results")
}

// TestScanStandaloneDrumkitXML verifies that standalone drumkit.xml files (not
// inside a tar archive) are discovered and parsed correctly.
func TestScanStandaloneDrumkitXML(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir)

	wantRelPaths := []string{
		"legacy-drumkits/kit-1.2.3/drumkit.xml",
		"legacy-drumkits/kit-0.9.3/drumkit.xml",
	}

	found := make(map[string]bool)
	for _, a := range artifacts {
		for _, want := range wantRelPaths {
			if a.RelPath == want {
				if _, ok := a.Metadata.(*model.DrumkitMetadata); !ok {
					t.Errorf("%s: got %T, want *model.DrumkitMetadata", a.RelPath, a.Metadata)
				}
				found[want] = true
			}
		}
	}
	for _, want := range wantRelPaths {
		if !found[want] {
			t.Errorf("standalone drumkit not found: %s", want)
		}
	}
}

// TestScanSongMetadata verifies that song files are parsed into *model.SongMetadata.
func TestScanSongMetadata(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir)

	for _, a := range artifacts {
		if a.RelPath == "v2.0.0.h2song" {
			if _, ok := a.Metadata.(*model.SongMetadata); !ok {
				t.Errorf("v2.0.0.h2song: got %T, want *model.SongMetadata", a.Metadata)
			}
			return
		}
	}
	t.Error("v2.0.0.h2song not found in scan results")
}
