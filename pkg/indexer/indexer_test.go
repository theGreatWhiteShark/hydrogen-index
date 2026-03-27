package indexer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/pkg/domain"
)

func TestBuildIndex(t *testing.T) {
	// Use a subset of res/hydrogen-artifacts
	// We'll create a temporary directory and copy some files there to have a controlled environment
	tempDir, err := os.MkdirTemp("", "hydrogen-index-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := "../../res/hydrogen-artifacts"
	filesToCopy := []string{
		"v2.0.0.h2pattern",
		"v2.0.0.h2song",
		"v2.0.0.h2drumkit",
	}

	for _, f := range filesToCopy {
		data, err := os.ReadFile(filepath.Join(srcDir, f))
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		err = os.WriteFile(filepath.Join(tempDir, f), data, 0644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", f, err)
		}
	}

	outputFile := filepath.Join(tempDir, "index.json")
	err = BuildIndex(tempDir, outputFile)
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Verify index.json exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("index.json was not generated")
	}

	// Read and parse index.json
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	var index domain.IndexFile
	err = json.Unmarshal(data, &index)
	if err != nil {
		t.Fatal(err)
	}

	if index.PatternCount != 1 {
		t.Errorf("expected 1 pattern, got %d", index.PatternCount)
	}
	if index.SongCount != 1 {
		t.Errorf("expected 1 song, got %d", index.SongCount)
	}
	if index.DrumkitCount != 1 {
		t.Errorf("expected 1 drumkit, got %d", index.DrumkitCount)
	}

	if index.Hash == "" {
		t.Error("hash should not be empty")
	}
}
