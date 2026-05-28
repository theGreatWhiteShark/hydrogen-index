package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
	"github.com/hydrogen-music/hydrogen-index/internal/scanner"
)

// sampleArtifacts returns a small, representative set of ArtifactFile values
// covering all three artifact types.
func sampleArtifacts() []scanner.ArtifactFile {
	return []scanner.ArtifactFile{
		{
			Path:    "/root/patterns/boom.h2pattern",
			RelPath: "patterns/boom.h2pattern",
			Hash:    "abc123",
			Size:    1024,
			Metadata: &model.PatternMetadata{
				Name:            "Boom",
				Author:          "Alice",
				Info:            "A simple boom pattern",
				License:         "CC0",
				FormatVersion:   2,
				UserVersion:     1,
				Tags:            []string{"rock"},
				Notes:           16,
				InstrumentTypes: []string{"kick"},
			},
		},
		{
			Path:    "/root/songs/mysong.h2song",
			RelPath: "songs/mysong.h2song",
			Hash:    "def456",
			Size:    2048,
			Metadata: &model.SongMetadata{
				Name:          "My Song",
				Author:        "Bob",
				Info:          "A demo song",
				License:       "CC-BY",
				FormatVersion: 3,
				UserVersion:   2,
				Tags:          []string{"demo"},
				Patterns:      4,
			},
		},
		{
			Path:    "/root/drumkits/kit.h2drumkit",
			RelPath: "drumkits/kit.h2drumkit",
			Hash:    "ghi789",
			Size:    8192,
			Metadata: &model.DrumkitMetadata{
				Name:            "Basic Kit",
				Author:          "Carol",
				Info:            "A basic drum kit",
				License:         "CC-BY-SA",
				FormatVersion:   2,
				UserVersion:     1,
				Tags:            []string{"acoustic"},
				FolderName:      "basic_kit",
				Instruments:     9,
				Components:      3,
				Samples:         27,
				InstrumentTypes: []string{"kick", "snare", "hihat"},
			},
		},
	}
}

func TestBuild_CountsAndTypes(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	if idx.PatternCount != 1 {
		t.Errorf("PatternCount = %d, want 1", idx.PatternCount)
	}
	if idx.SongCount != 1 {
		t.Errorf("SongCount = %d, want 1", idx.SongCount)
	}
	if idx.DrumkitCount != 1 {
		t.Errorf("DrumkitCount = %d, want 1", idx.DrumkitCount)
	}
	if len(idx.Patterns) != 1 {
		t.Fatalf("len(Patterns) = %d, want 1", len(idx.Patterns))
	}
	if len(idx.Songs) != 1 {
		t.Fatalf("len(Songs) = %d, want 1", len(idx.Songs))
	}
	if len(idx.Drumkits) != 1 {
		t.Fatalf("len(Drumkits) = %d, want 1", len(idx.Drumkits))
	}
}

func TestBuild_PatternFields(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	p := idx.Patterns[0]
	if p.Name != "Boom" {
		t.Errorf("Name = %q, want %q", p.Name, "Boom")
	}
	if p.URL != "patterns/boom.h2pattern" {
		t.Errorf("URL = %q, want %q", p.URL, "patterns/boom.h2pattern")
	}
	if p.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", p.Hash, "abc123")
	}
	if p.Author != "Alice" {
		t.Errorf("Author = %q, want Alice", p.Author)
	}
	if p.Description != "A simple boom pattern" {
		t.Errorf("Description = %q", p.Description)
	}
	if p.Notes != 16 {
		t.Errorf("Notes = %d, want 16", p.Notes)
	}
	if p.Size != 1024 {
		t.Errorf("Size = %d, want 1024", p.Size)
	}
	if p.Type != model.ArtifactTypePattern {
		t.Errorf("Type = %q, want %q", p.Type, model.ArtifactTypePattern)
	}
	if p.InstrumentTypes == nil {
		t.Error("InstrumentTypes must not be nil")
	}
}

func TestBuild_SongFields(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	s := idx.Songs[0]
	if s.Name != "My Song" {
		t.Errorf("Name = %q, want %q", s.Name, "My Song")
	}
	if s.Patterns != 4 {
		t.Errorf("Patterns = %d, want 4", s.Patterns)
	}
	if s.Type != model.ArtifactTypeSong {
		t.Errorf("Type = %q, want %q", s.Type, model.ArtifactTypeSong)
	}
	if s.Tags == nil {
		t.Error("Tags must not be nil")
	}
}

func TestBuild_DrumkitFields(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	d := idx.Drumkits[0]
	if d.Name != "Basic Kit" {
		t.Errorf("Name = %q, want %q", d.Name, "Basic Kit")
	}
	if d.FolderName != "basic_kit" {
		t.Errorf("FolderName = %q, want %q", d.FolderName, "basic_kit")
	}
	if d.Instruments != 9 {
		t.Errorf("Instruments = %d, want 9", d.Instruments)
	}
	if d.Components != 3 {
		t.Errorf("Components = %d, want 3", d.Components)
	}
	if d.Samples != 27 {
		t.Errorf("Samples = %d, want 27", d.Samples)
	}
	if d.Type != model.ArtifactTypeDrumkit {
		t.Errorf("Type = %q, want %q", d.Type, model.ArtifactTypeDrumkit)
	}
	if d.InstrumentTypes == nil {
		t.Error("InstrumentTypes must not be nil")
	}
}

// TestBuild_NilSlicesSafe verifies that nil slice fields in metadata are
// promoted to empty (non-nil) slices so the JSON output never contains null.
func TestBuild_NilSlicesSafe(t *testing.T) {
	artifacts := []scanner.ArtifactFile{
		{
			RelPath:  "p.h2pattern",
			Hash:     "h",
			Metadata: &model.PatternMetadata{Name: "P"},
		},
		{
			RelPath:  "d.h2drumkit",
			Hash:     "h",
			Metadata: &model.DrumkitMetadata{Name: "D"},
		},
	}

	idx, err := Build(artifacts)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if idx.Patterns[0].Tags == nil {
		t.Error("pattern Tags must not be nil when source is nil")
	}
	if idx.Patterns[0].InstrumentTypes == nil {
		t.Error("pattern InstrumentTypes must not be nil when source is nil")
	}
	if idx.Drumkits[0].Tags == nil {
		t.Error("drumkit Tags must not be nil when source is nil")
	}
	if idx.Drumkits[0].InstrumentTypes == nil {
		t.Error("drumkit InstrumentTypes must not be nil when source is nil")
	}
}

func TestBuild_UnknownMetadataType(t *testing.T) {
	artifacts := []scanner.ArtifactFile{
		{RelPath: "x", Hash: "h", Metadata: "not a known type"},
	}
	_, err := Build(artifacts)
	if err == nil {
		t.Fatal("expected error for unknown metadata type, got nil")
	}
}

func TestFinalize_HashPopulated(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := Finalize(idx)
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	if !strings.Contains(string(data), `"hash"`) {
		t.Error("JSON output does not contain a hash field")
	}
	if idx.Hash == "" {
		t.Error("idx.Hash is empty after Finalize")
	}
}

// TestFinalize_HashMatchesCompact verifies that the hash computed by Finalize
// is SHA-256 of the canonical compact JSON with the "hash" key removed.
// Canonical form uses alphabetically sorted keys at all nesting levels,
// matching the C++ parser in OnlineImporter.cpp which parses the JSON into
// QJsonObject and re-serializes with QJsonDocument::Compact.
func TestFinalize_HashMatchesCompact(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := Finalize(idx)
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	// Parse the finalized JSON
	var parsed model.Index
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	gotHash := parsed.Hash

	// Compute expected hash using the same canonical approach: marshal to
	// map[string]interface{} for sorted keys, then remove "hash" key.
	parsed.Hash = ""
	compact, err := json.Marshal(&parsed)
	if err != nil {
		t.Fatalf("compact marshal: %v", err)
	}

	var canonical map[string]interface{}
	if err := json.Unmarshal(compact, &canonical); err != nil {
		t.Fatalf("unmarshal for canonical: %v", err)
	}
	delete(canonical, "hash")

	canonicalBytes, err := json.Marshal(canonical)
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}

	expectedHash := sha256.Sum256(canonicalBytes)
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	if gotHash != expectedHashHex {
		t.Errorf("hash mismatch:\n  got:      %s\n  expected: %s\n  canonical: %s", gotHash, expectedHashHex, canonicalBytes)
	}
}

func TestFinalize_Roundtrip(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := Finalize(idx)
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	var result model.Index
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result.PatternCount != 1 {
		t.Errorf("PatternCount = %d, want 1", result.PatternCount)
	}
	if result.SongCount != 1 {
		t.Errorf("SongCount = %d, want 1", result.SongCount)
	}
	if result.DrumkitCount != 1 {
		t.Errorf("DrumkitCount = %d, want 1", result.DrumkitCount)
	}

	// SHA-256 hex digest is always exactly 64 characters.
	if len(result.Hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(result.Hash))
	}
}

// TestBuild_DuplicateNames verifies that Build does not error on duplicate
// names; validation of uniqueness is deferred to a future pass.
func TestBuild_DuplicateNames(t *testing.T) {
	artifacts := []scanner.ArtifactFile{
		{
			RelPath:  "a.h2pattern",
			Hash:     "h1",
			Metadata: &model.PatternMetadata{Name: "Same"},
		},
		{
			RelPath:  "b.h2pattern",
			Hash:     "h2",
			Metadata: &model.PatternMetadata{Name: "Same"},
		},
	}
	_, err := Build(artifacts)
	if err != nil {
		t.Fatalf("Build should not error on duplicate names (future validation): %v", err)
	}
}

func TestBuild_EmptyArtifacts(t *testing.T) {
	idx, err := Build(nil)
	if err != nil {
		t.Fatalf("Build(nil): %v", err)
	}
	if idx.PatternCount != 0 || idx.SongCount != 0 || idx.DrumkitCount != 0 {
		t.Error("counts must all be zero for empty input")
	}
	if idx.Patterns == nil || idx.Songs == nil || idx.Drumkits == nil {
		t.Error("slices must be non-nil (not null in JSON)")
	}
}

func TestBuild_IndexVersion(t *testing.T) {
	idx, err := Build(nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if idx.Version != model.Version {
		t.Errorf("Version = %q, want %q", idx.Version, model.Version)
	}
	if idx.Created == "" {
		t.Error("Created must not be empty")
	}
}

// TestBuild_URLWithBaseURL verifies that BaseURL is prepended to RelPath.
func TestBuild_URLWithBaseURL(t *testing.T) {
	artifacts := []scanner.ArtifactFile{
		{
			RelPath: "patterns/boom.h2pattern",
			BaseURL: "https://raw.githubusercontent.com/user/repo/main",
			Hash:    "abc",
			Metadata: &model.PatternMetadata{
				Name: "Boom",
				Tags: []string{},
			},
		},
	}
	idx, err := Build(artifacts)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := "https://raw.githubusercontent.com/user/repo/main/patterns/boom.h2pattern"
	if idx.Patterns[0].URL != want {
		t.Errorf("URL = %q, want %q", idx.Patterns[0].URL, want)
	}
}

// TestBuild_URLWithoutBaseURL verifies that empty BaseURL yields RelPath only.
func TestBuild_URLWithoutBaseURL(t *testing.T) {
	artifacts := []scanner.ArtifactFile{
		{
			RelPath: "drumkits/kit.h2drumkit",
			BaseURL: "",
			Hash:    "def",
			Metadata: &model.DrumkitMetadata{
				Name: "Kit",
				Tags: []string{},
			},
		},
	}
	idx, err := Build(artifacts)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := "drumkits/kit.h2drumkit"
	if idx.Drumkits[0].URL != want {
		t.Errorf("URL = %q, want %q", idx.Drumkits[0].URL, want)
	}
}

// TestFinalize_CrossPlatformHash verifies that the hash computed by Finalize
// matches what the C++ OnlineImporter.cpp parser would compute when it:
//  1. Parses the finalized JSON into QJsonObject
//  2. Removes the "hash" key
//  3. Re-serializes with QJsonDocument::Compact
//
// Qt 5 QJsonObject iterates keys in alphabetical order. Qt 6 preserves
// insertion order — but since our canonical form writes sorted keys,
// parsing and re-serializing in Qt 6 also yields sorted keys.
//
// This test simulates the C++ behavior by unmarshaling into
// map[string]interface{} (which sorts keys alphabetically when marshaling
// back), then hashing — exactly what the C++ side does.
func TestFinalize_CrossPlatformHash(t *testing.T) {
	idx, err := Build(sampleArtifacts())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := Finalize(idx)
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	// Simulate C++ OnlineImporter::parseIndex behavior:
	// 1. Parse JSON into QJsonObject (map[string]interface{} in Go)
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("unmarshal finalized JSON: %v", err)
	}

	// 2. Extract expected hash from top-level
	sExpectedHash, ok := root["hash"].(string)
	if !ok || sExpectedHash == "" {
		t.Fatal("expected 'hash' field in finalized JSON")
	}

	// 3. Remove "hash" key (QJsonObject::remove("hash"))
	delete(root, "hash")

	// 4. Re-serialize to compact JSON (QJsonDocument::Compact)
	// map[string]interface{} marshals with sorted keys
	dataWithoutHash, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshal without hash: %v", err)
	}

	// 5. Compute SHA-256 of compact JSON
	sComputedHash := sha256.Sum256(dataWithoutHash)
	sComputedHashHex := hex.EncodeToString(sComputedHash[:])

	// 6. Verify match (this is the assertion that would fail before the fix)
	if sComputedHashHex != sExpectedHash {
		t.Errorf("Cross-platform hash mismatch (simulating C++ QJsonObject behavior):\n"+
			"  Expected (from index): %s\n"+
			"  Computed (sorted keys): %s\n"+
			"  Compact JSON: %s",
			sExpectedHash, sComputedHashHex, dataWithoutHash)
	}
}
