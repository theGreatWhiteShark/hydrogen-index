package domain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestUnmarshalIndexFile(t *testing.T) {
	data, err := os.ReadFile("../../res/references-index.json")
	if err != nil {
		t.Fatalf("failed to read reference file: %v", err)
	}

	var index IndexFile
	err = json.Unmarshal(data, &index)
	if err != nil {
		t.Fatalf("failed to unmarshal index file: %v", err)
	}

	if index.PatternCount != 1 {
		t.Errorf("expected patternCount 1, got %d", index.PatternCount)
	}

	if len(index.Patterns) == 0 {
		t.Fatal("expected at least one pattern")
	}

	pattern := index.Patterns[0]
	if pattern.Name != "Example Pattern" {
		t.Errorf("expected pattern name 'Example Pattern', got '%s'", pattern.Name)
	}

	if pattern.Notes != 35 {
		t.Errorf("expected 35 notes, got %d", pattern.Notes)
	}

	if pattern.Type != "pattern" {
		t.Errorf("expected type 'pattern', got '%s'", pattern.Type)
	}

	if pattern.Author != "Name" {
		t.Errorf("expected author 'Name', got '%s'", pattern.Author)
	}

	if len(pattern.InstrumentTypes) != 3 {
		t.Errorf("expected 3 instrument types, got %d", len(pattern.InstrumentTypes))
	}

	if index.SongCount != 1 {
		t.Errorf("expected songCount 1, got %d", index.SongCount)
	}

	if index.DrumkitCount != 1 {
		t.Errorf("expected drumkitCount 1, got %d", index.DrumkitCount)
	}

	if index.Hash == "" {
		t.Error("expected hash to be populated")
	}
}
