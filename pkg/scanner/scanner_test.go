package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitroot-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	gitRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(filepath.Join(gitRoot, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	deepDir := filepath.Join(gitRoot, "a", "b", "c")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := FindGitRoot(deepDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, _ := filepath.Abs(gitRoot)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	_, err = FindGitRoot(tmpDir)
	if err == nil {
		t.Error("expected error for non-git root path")
	}
}

func TestScanArtifacts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := []string{
		"a.h2pattern",
		"b.h2song",
		"c.h2drumkit",
		"d.txt",
		"subdir/e.h2pattern",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ScanArtifacts(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantCount := 4
	if len(got) != wantCount {
		t.Errorf("got %d artifacts, want %d", len(got), wantCount)
	}
}

func TestExtractDrumkitXML(t *testing.T) {
	// The prompt mentions res/hydrogen-artifacts/v2.0.0.h2drumkit
	// We should use an absolute path to ensure the test can find it from the package directory.
	// However, we'll try relative path first, as `go test` runs from the package directory.
	
	// First, find the project root to construct the path to the artifact.
	root, err := FindGitRoot(".")
	if err != nil {
		t.Fatalf("could not find git root: %v", err)
	}
	
	artifactPath := filepath.Join(root, "res", "hydrogen-artifacts", "v2.0.0.h2drumkit")
	
	// Check if artifact exists
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Skipf("skipping TestExtractDrumkitXML: artifact not found at %s", artifactPath)
	}

	tmpXML, err := ExtractDrumkitXML(artifactPath)
	if err != nil {
		t.Fatalf("failed to extract drumkit.xml: %v", err)
	}
	defer os.Remove(tmpXML)

	content, err := os.ReadFile(tmpXML)
	if err != nil {
		t.Fatalf("failed to read extracted XML: %v", err)
	}

	if len(content) == 0 {
		t.Error("extracted XML is empty")
	}
}
