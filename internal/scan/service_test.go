package scan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
	internalvalidate "github.com/theGreatWhiteShark/hydrogen-index/internal/validate"
)

func TestRunWritesHashedIndexForMixedArtifacts(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2song",
		filepath.Join(repoRoot, "songs", "v2.0.0.h2song"),
	)
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-songs/test_song_1.2.2.h2song",
		filepath.Join(repoRoot, "songs", "legacy", "test_song_1.2.2.h2song"),
	)
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2pattern",
		filepath.Join(repoRoot, "patterns", "v2.0.0.h2pattern"),
	)
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-patterns/legacy_pattern.h2pattern",
		filepath.Join(repoRoot, "patterns", "legacy_pattern.h2pattern"),
	)
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2drumkit",
		filepath.Join(repoRoot, "drumkits", "v2.0.0.h2drumkit"),
	)
	mustCopyFile(t,
		"/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-drumkits/kit-1.2.3/drumkit.xml",
		filepath.Join(repoRoot, "drumkits", "legacy", "drumkit.xml"),
	)

	workingDir := filepath.Join(repoRoot, "nested", "workspace")
	mustMkdirAll(t, workingDir)
	outputPath := filepath.Join(workingDir, "index.json")

	if err := Run(Options{
		WorkingDir: workingDir,
		OutputPath: outputPath,
		BaseURL:    "https://example.com/repository",
		Version:    "0.1.0",
		Now:        func() time.Time { return time.Date(2026, 3, 27, 14, 0, 0, 0, time.UTC) },
	}); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) failed: %v", outputPath, err)
	}

	var document indexfile.Document
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if document.PatternCount != 2 || document.SongCount != 2 || document.DrumkitCount != 2 {
		t.Fatalf("unexpected top-level counts: %+v", document)
	}

	if document.Hash == "" {
		t.Fatalf("expected top-level hash to be populated: %+v", document)
	}

	if err := internalvalidate.ValidateBytes(data); err != nil {
		t.Fatalf("generated index did not validate: %v", err)
	}
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
