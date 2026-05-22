package scanner_test

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
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

// TestScanDrumkitTarFolderName verifies that the folderName field is extracted
// from the top-level folder inside a .h2drumkit tar archive.
func TestScanDrumkitTarFolderName(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir)

	for _, a := range artifacts {
		if a.RelPath == "v2.0.0.h2drumkit" {
			meta, ok := a.Metadata.(*model.DrumkitMetadata)
			if !ok {
				t.Fatalf("v2.0.0.h2drumkit: got %T, want *model.DrumkitMetadata", a.Metadata)
			}
			if meta.FolderName != "testKit" {
				t.Errorf("v2.0.0.h2drumkit FolderName = %q, want %q", meta.FolderName, "testKit")
			}
			return
		}
	}
	t.Error("v2.0.0.h2drumkit not found in scan results")
}

// TestScanDrumkitTarNoTopLevelFolder verifies that an archive with root-level
// files (no containing folder) produces an error.
func TestScanDrumkitTarNoTopLevelFolder(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.h2drumkit")

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	// Write drumkit.xml directly at root (no folder).
	hdr := &tar.Header{
		Name: "drumkit.xml",
		Mode: 0o644,
		Size: int64(len(drumkitXML)),
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := w.Write([]byte(drumkitXML)); err != nil {
		t.Fatalf("write data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, errs := scanner.Scan(tmpDir)
	if len(errs) == 0 {
		t.Fatal("expected error for archive with no top-level folder, got none")
	}
}

// TestScanDrumkitTarMultipleTopLevelFolders verifies that an archive with
// multiple top-level folders produces an error.
func TestScanDrumkitTarMultipleTopLevelFolders(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.h2drumkit")

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	// Write a directory entry for folder1.
	if err := w.WriteHeader(&tar.Header{Name: "folder1/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write drumkit.xml inside folder1.
	hdr := &tar.Header{
		Name: "folder1/drumkit.xml",
		Mode: 0o644,
		Size: int64(len(drumkitXML)),
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := w.Write([]byte(drumkitXML)); err != nil {
		t.Fatalf("write data: %v", err)
	}
	// Write a directory entry for folder2.
	if err := w.WriteHeader(&tar.Header{Name: "folder2/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write a dummy file inside folder2.
	if err := w.WriteHeader(&tar.Header{Name: "folder2/readme.txt", Mode: 0o644, Size: 4}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := io.WriteString(w, "test"); err != nil {
		t.Fatalf("write data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, errs := scanner.Scan(tmpDir)
	if len(errs) == 0 {
		t.Fatal("expected error for archive with multiple top-level folders, got none")
	}
}

// drumkitXML is a minimal valid drumkit.xml for test fixtures.
const drumkitXML = `<?xml version="1.0"?>
<drumkit_info>
  <name>Test</name>
  <formatVersion>2</formatVersion>
  <instrumentList/>
</drumkit_info>`
