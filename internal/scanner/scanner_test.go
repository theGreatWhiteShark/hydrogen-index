package scanner_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	artifacts, errs := scanner.Scan(dir, "", nil)

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
	if got, want := len(drumkits), 2; got != want {
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
	artifacts, _ := scanner.Scan(dir, "", nil)

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
	artifacts, _ := scanner.Scan(dir, "", nil)

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

// TestScanSongMetadata verifies that song files are parsed into *model.SongMetadata.
func TestScanSongMetadata(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, _ := scanner.Scan(dir, "", nil)

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
	artifacts, _ := scanner.Scan(dir, "", nil)

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

	_, errs := scanner.Scan(tmpDir, "", nil)
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

	_, errs := scanner.Scan(tmpDir, "", nil)
	if len(errs) == 0 {
		t.Fatal("expected error for archive with multiple top-level folders, got none")
	}
}

// TestScanDrumkitTarGzip verifies that gzip-compressed .h2drumkit files
// are correctly parsed.
func TestScanDrumkitTarGzip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "compressed.h2drumkit")

	// Create a tar archive.
	var tarBuf bytes.Buffer
	w := tar.NewWriter(&tarBuf)
	// Write a directory entry.
	if err := w.WriteHeader(&tar.Header{Name: "testKit/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write drumkit.xml inside the folder.
	hdr := &tar.Header{
		Name: "testKit/drumkit.xml",
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

	// Compress the tar archive with gzip.
	var gzipBuf bytes.Buffer
	gz := gzip.NewWriter(&gzipBuf)
	if _, err := gz.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	if err := os.WriteFile(path, gzipBuf.Bytes(), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artifacts, errs := scanner.Scan(tmpDir, "", nil)
	if len(errs) != 0 {
		t.Fatalf("scan errors: %v", errs)
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifact count: got %d, want 1", len(artifacts))
	}

	meta, ok := artifacts[0].Metadata.(*model.DrumkitMetadata)
	if !ok {
		t.Fatalf("metadata type: got %T, want *model.DrumkitMetadata", artifacts[0].Metadata)
	}
	if meta.FolderName != "testKit" {
		t.Errorf("FolderName = %q, want %q", meta.FolderName, "testKit")
	}
}

// TestScanDrumkitTarWithIgnoredFiles verifies that archives with ignored
// auxiliary files at the top level are still parsed correctly.
func TestScanDrumkitTarWithIgnoredFiles(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "with-ignored.h2drumkit")

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)

	// Write ignored macOS Apple Double file at root level.
	if err := w.WriteHeader(&tar.Header{Name: "._testKit", Mode: 0o644, Size: 0}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write .DS_Store at root level.
	if err := w.WriteHeader(&tar.Header{Name: ".DS_Store", Mode: 0o644, Size: 0}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write Thumbs.db at root level.
	if err := w.WriteHeader(&tar.Header{Name: "Thumbs.db", Mode: 0o644, Size: 0}); err != nil {
		t.Fatalf("write header: %v", err)
	}

	// Write the actual drumkit folder.
	if err := w.WriteHeader(&tar.Header{Name: "testKit/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write drumkit.xml inside the folder.
	hdr := &tar.Header{
		Name: "testKit/drumkit.xml",
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

	artifacts, errs := scanner.Scan(tmpDir, "", nil)
	if len(errs) != 0 {
		t.Fatalf("scan errors: %v", errs)
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifact count: got %d, want 1", len(artifacts))
	}

	meta, ok := artifacts[0].Metadata.(*model.DrumkitMetadata)
	if !ok {
		t.Fatalf("metadata type: got %T, want *model.DrumkitMetadata", artifacts[0].Metadata)
	}
	if meta.FolderName != "testKit" {
		t.Errorf("FolderName = %q, want %q", meta.FolderName, "testKit")
	}
}

// TestScanDrumkitTarWithIgnoredFolders verifies that archives with ignored
// auxiliary folders at the top level are still parsed correctly.
func TestScanDrumkitTarWithIgnoredFolders(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "with-ignored-folders.h2drumkit")

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)

	// Write ignored .git folder at root level.
	if err := w.WriteHeader(&tar.Header{Name: ".git/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write ignored node_modules folder at root level.
	if err := w.WriteHeader(&tar.Header{Name: "node_modules/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}

	// Write the actual drumkit folder.
	if err := w.WriteHeader(&tar.Header{Name: "testKit/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write drumkit.xml inside the folder.
	hdr := &tar.Header{
		Name: "testKit/drumkit.xml",
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

	artifacts, errs := scanner.Scan(tmpDir, "", nil)
	if len(errs) != 0 {
		t.Fatalf("scan errors: %v", errs)
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifact count: got %d, want 1", len(artifacts))
	}

	meta, ok := artifacts[0].Metadata.(*model.DrumkitMetadata)
	if !ok {
		t.Fatalf("metadata type: got %T, want *model.DrumkitMetadata", artifacts[0].Metadata)
	}
	if meta.FolderName != "testKit" {
		t.Errorf("FolderName = %q, want %q", meta.FolderName, "testKit")
	}
}

// TestScanDrumkitTarWithNonIgnoredTopLevelFile verifies that archives with
// non-ignored files at the top level still produce an error.
func TestScanDrumkitTarWithNonIgnoredTopLevelFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.h2drumkit")

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)

	// Write a non-ignored file at root level.
	if err := w.WriteHeader(&tar.Header{Name: "readme.txt", Mode: 0o644, Size: 4}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := io.WriteString(w, "test"); err != nil {
		t.Fatalf("write data: %v", err)
	}

	// Write the actual drumkit folder.
	if err := w.WriteHeader(&tar.Header{Name: "testKit/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	// Write drumkit.xml inside the folder.
	hdr := &tar.Header{
		Name: "testKit/drumkit.xml",
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

	_, errs := scanner.Scan(tmpDir, "", nil)
	if len(errs) == 0 {
		t.Fatal("expected error for archive with non-ignored top-level file, got none")
	}
}

// TestScanBaseURL verifies that the BaseURL is set on all artifacts when
// provided to Scan.
func TestScanBaseURL(t *testing.T) {
	dir := artifactsDir(t)
	baseURL := "https://example.com/artifacts/main"
	artifacts, errs := scanner.Scan(dir, baseURL, nil)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}
	for _, a := range artifacts {
		if a.BaseURL != baseURL {
			t.Errorf("%s: BaseURL = %q, want %q", a.RelPath, a.BaseURL, baseURL)
		}
	}
}

// TestScanEmptyBaseURL verifies that empty BaseURL is left as-is.
func TestScanEmptyBaseURL(t *testing.T) {
	dir := artifactsDir(t)
	artifacts, errs := scanner.Scan(dir, "", nil)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}
	for _, a := range artifacts {
		if a.BaseURL != "" {
			t.Errorf("%s: BaseURL = %q, want empty", a.RelPath, a.BaseURL)
		}
	}
}

// drumkitXML is a minimal valid drumkit.xml for test fixtures.
const drumkitXML = `<?xml version="1.0"?>
<drumkit_info>
  <name>Test</name>
  <formatVersion>2</formatVersion>
  <instrumentList/>
</drumkit_info>`

// TestScanExcludeSingleDirectory verifies that a single directory can be excluded
// from scanning by its name.
func TestScanExcludeSingleDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure with artifacts
	excludedDir := filepath.Join(tmpDir, "test")
	if err := os.Mkdir(excludedDir, 0o755); err != nil {
		t.Fatalf("create excluded dir: %v", err)
	}

	// Create an artifact in the excluded directory
	excludedArtifact := filepath.Join(excludedDir, "excluded.h2pattern")
	if err := os.WriteFile(excludedArtifact, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write excluded artifact: %v", err)
	}

	// Create an artifact in the root directory (should be scanned)
	includedArtifact := filepath.Join(tmpDir, "included.h2pattern")
	if err := os.WriteFile(includedArtifact, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write included artifact: %v", err)
	}

	// Scan with the directory excluded
	artifacts, errs := scanner.Scan(tmpDir, "", []string{"test"})
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// Verify only the included artifact was found
	if got, want := len(artifacts), 1; got != want {
		t.Errorf("artifact count: got %d, want %d", got, want)
	}

	// Verify the excluded artifact was not found
	for _, a := range artifacts {
		if a.RelPath == "test/excluded.h2pattern" {
			t.Errorf("excluded artifact was scanned: %s", a.RelPath)
		}
	}
}

// TestScanExcludeMultipleDirectories verifies that multiple directories can be
// excluded from scanning.
func TestScanExcludeMultipleDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple directories to exclude
	excludedDirs := []string{"test", "node_modules", ".git"}
	for _, dirName := range excludedDirs {
		dirPath := filepath.Join(tmpDir, dirName)
		if err := os.Mkdir(dirPath, 0o755); err != nil {
			t.Fatalf("create excluded dir %s: %v", dirName, err)
		}

		// Add an artifact to each excluded directory
		artifactPath := filepath.Join(dirPath, dirName+".h2pattern")
		if err := os.WriteFile(artifactPath, []byte(patternXML), 0o644); err != nil {
			t.Fatalf("write artifact in %s: %v", dirName, err)
		}
	}

	// Create an artifact in the root directory (should be scanned)
	includedArtifact := filepath.Join(tmpDir, "included.h2pattern")
	if err := os.WriteFile(includedArtifact, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write included artifact: %v", err)
	}

	// Scan with all directories excluded
	artifacts, errs := scanner.Scan(tmpDir, "", excludedDirs)
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// Verify only the included artifact was found
	if got, want := len(artifacts), 1; got != want {
		t.Errorf("artifact count: got %d, want %d", got, want)
	}

	// Verify no excluded artifacts were found
	for _, a := range artifacts {
		for _, dirName := range excludedDirs {
			if strings.HasPrefix(a.RelPath, dirName+"/") {
				t.Errorf("artifact in excluded directory was scanned: %s", a.RelPath)
			}
		}
	}
}

// TestScanExcludeRelativePath verifies that directories can be excluded by
// their relative path from the scan root.
func TestScanExcludeRelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested directory structure
	nestedDir := filepath.Join(tmpDir, "res", "test-artifacts")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}

	// Create an artifact in the nested directory
	excludedArtifact := filepath.Join(nestedDir, "excluded.h2pattern")
	if err := os.WriteFile(excludedArtifact, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write excluded artifact: %v", err)
	}

	// Create an artifact in the root directory (should be scanned)
	includedArtifact := filepath.Join(tmpDir, "included.h2pattern")
	if err := os.WriteFile(includedArtifact, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write included artifact: %v", err)
	}

	// Scan with the relative path excluded
	artifacts, errs := scanner.Scan(tmpDir, "", []string{"res/test-artifacts"})
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// Verify only the included artifact was found
	if got, want := len(artifacts), 1; got != want {
		t.Errorf("artifact count: got %d, want %d", got, want)
	}

	// Verify the excluded artifact was not found
	for _, a := range artifacts {
		if a.RelPath == "res/test-artifacts/excluded.h2pattern" {
			t.Errorf("excluded artifact was scanned: %s", a.RelPath)
		}
	}
}

// TestScanExcludeEmpty verifies that an empty exclude list does not affect scanning.
func TestScanExcludeEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an artifact
	artifactPath := filepath.Join(tmpDir, "test.h2pattern")
	if err := os.WriteFile(artifactPath, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	// Scan with empty exclude list
	artifacts, errs := scanner.Scan(tmpDir, "", []string{})
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// Verify the artifact was found
	if got, want := len(artifacts), 1; got != want {
		t.Errorf("artifact count: got %d, want %d", got, want)
	}
}

// TestScanExcludeNonExistent verifies that excluding non-existent paths does not
// cause errors.
func TestScanExcludeNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an artifact
	artifactPath := filepath.Join(tmpDir, "test.h2pattern")
	if err := os.WriteFile(artifactPath, []byte(patternXML), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	// Scan with non-existent exclude paths
	artifacts, errs := scanner.Scan(tmpDir, "", []string{"nonexistent", "also-nonexistent"})
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("scan error: %v", e)
		}
	}

	// Verify the artifact was still found
	if got, want := len(artifacts), 1; got != want {
		t.Errorf("artifact count: got %d, want %d", got, want)
	}
}

// patternXML is a minimal valid pattern XML for test fixtures.
const patternXML = `<?xml version="1.0"?>
<drumkit_pattern>
  <pattern_name>Test</pattern_name>
  <version>2.0.0</version>
  <author>Test Author</author>
  <info>Test pattern</info>
  <license>GPL</license>
</drumkit_pattern>`
