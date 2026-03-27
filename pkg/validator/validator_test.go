package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateIndexFile(t *testing.T) {
	// 1. Success case: Validating the reference file
	// Note: We need to use the absolute path or a reliable relative path to res/references-index.json
	refFile := "../../res/references-index.json"
	err := ValidateIndexFile(refFile)
	if err != nil {
		t.Errorf("expected %s to be valid, got error: %v", refFile, err)
	}

	// 2. Failure case: Missing hash in a pattern
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	
	// Broken JSON: missing 'hash' in patterns[0]
	brokenJSON := `{
  "version": "0.1.0",
  "created": "2026-03-24T21:20:00",
  "patternCount": 1,
  "songCount": 0,
  "drumkitCount": 0,
  "patterns": [
    {
      "type": "pattern",
      "name": "Example Pattern",
      "url": "permalink",
      "author": "Name",
      "description": "Missing hash",
      "version": 1,
      "formatVersion": 1,
      "tags": [],
      "size": 3564,
      "license": "CC0",
      "notes": 35,
      "instrumentTypes": []
    }
  ],
  "songs": [],
  "drumkits": [],
  "hash": "somehash"
}`
	if err := os.WriteFile(invalidFile, []byte(brokenJSON), 0644); err != nil {
		t.Fatal(err)
	}

	err = ValidateIndexFile(invalidFile)
	if err == nil {
		t.Error("expected validation to fail for missing hash, but it passed")
	}

	// 3. Failure case: Wrong type for 'size'
	wrongTypeJSON := `{
  "version": "0.1.0",
  "created": "2026-03-24T21:20:00",
  "patternCount": 0,
  "songCount": 1,
  "drumkitCount": 0,
  "patterns": [],
  "songs": [
    {
      "type": "song",
      "name": "Example Song",
      "url": "permalink",
      "hash": "hash",
      "author": "Name",
      "description": "Wrong type for size",
      "version": 4,
      "formatVersion": 1,
      "tags": [],
      "size": "not-a-number",
      "license": "CC0",
      "patterns": 10
    }
  ],
  "drumkits": [],
  "hash": "somehash"
}`
	wrongTypeFile := filepath.Join(tmpDir, "wrong_type.json")
	if err := os.WriteFile(wrongTypeFile, []byte(wrongTypeJSON), 0644); err != nil {
		t.Fatal(err)
	}

	err = ValidateIndexFile(wrongTypeFile)
	if err == nil {
		t.Error("expected validation to fail for wrong type of size, but it passed")
	}
}
