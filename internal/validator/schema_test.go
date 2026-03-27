package validator_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/validator"
)

// TestValidateReference confirms the canonical reference file passes validation.
func TestValidateReference(t *testing.T) {
	if err := validator.Validate("../../res/references-index.json"); err != nil {
		t.Fatalf("expected valid, got error: %v", err)
	}
}

// TestValidateMissingRequiredField checks that an index missing a required
// top-level field is correctly rejected.
func TestValidateMissingRequiredField(t *testing.T) {
	// "version" is absent.
	const input = `{
		"created": "2026-01-01T00:00:00",
		"patternCount": 0,
		"songCount": 0,
		"drumkitCount": 0,
		"patterns": [],
		"songs": [],
		"drumkits": []
	}`
	path := writeTemp(t, input)
	if err := validator.Validate(path); err == nil {
		t.Fatal("expected error for missing 'version' field, got nil")
	}
}

// TestValidateMalformedJSON checks that non-JSON input returns a parse error.
func TestValidateMalformedJSON(t *testing.T) {
	path := writeTemp(t, `{not valid json}`)
	err := validator.Validate(path)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse error message, got: %v", err)
	}
}

// TestValidateNegativeCount checks that a negative count value is rejected.
func TestValidateNegativeCount(t *testing.T) {
	const input = `{
		"version": "0.1.0",
		"created": "2026-01-01T00:00:00",
		"patternCount": -1,
		"songCount": 0,
		"drumkitCount": 0,
		"patterns": [],
		"songs": [],
		"drumkits": []
	}`
	path := writeTemp(t, input)
	if err := validator.Validate(path); err == nil {
		t.Fatal("expected error for negative patternCount, got nil")
	}
}

// TestValidateWrongFieldType checks that a wrong type for a field is rejected.
func TestValidateWrongFieldType(t *testing.T) {
	// "version" must be a string, not a number.
	const input = `{
		"version": 1,
		"created": "2026-01-01T00:00:00",
		"patternCount": 0,
		"songCount": 0,
		"drumkitCount": 0,
		"patterns": [],
		"songs": [],
		"drumkits": []
	}`
	path := writeTemp(t, input)
	if err := validator.Validate(path); err == nil {
		t.Fatal("expected error for integer 'version', got nil")
	}
}

// TestValidateMissingPatternsArray checks that omitting the patterns array
// is rejected even when patternCount is zero.
func TestValidateMissingPatternsArray(t *testing.T) {
	const input = `{
		"version": "0.1.0",
		"created": "2026-01-01T00:00:00",
		"patternCount": 0,
		"songCount": 0,
		"drumkitCount": 0,
		"songs": [],
		"drumkits": []
	}`
	path := writeTemp(t, input)
	if err := validator.Validate(path); err == nil {
		t.Fatal("expected error for missing 'patterns' array, got nil")
	}
}

// TestValidatePatternMissingRequired checks that a pattern entry missing a
// required field (notes) is rejected.
func TestValidatePatternMissingRequired(t *testing.T) {
	const input = `{
		"version": "0.1.0",
		"created": "2026-01-01T00:00:00",
		"patternCount": 1,
		"songCount": 0,
		"drumkitCount": 0,
		"patterns": [{
			"type": "pattern",
			"name": "Test",
			"url": "http://example.com",
			"hash": "abc123",
			"author": "Author",
			"description": "Desc",
			"version": 1,
			"formatVersion": 1,
			"tags": [],
			"size": 100,
			"license": "CC0",
			"instrumentTypes": []
		}],
		"songs": [],
		"drumkits": []
	}`
	path := writeTemp(t, input)
	if err := validator.Validate(path); err == nil {
		t.Fatal("expected error for pattern missing 'notes', got nil")
	}
}

// writeTemp writes content to a temporary file and returns its path.
// The file is removed when the test ends.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "index-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}
