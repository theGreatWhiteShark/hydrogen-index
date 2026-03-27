package parser_test

import (
	"os"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/parser"
)

// TestParsePattern exercises all three .h2pattern format variants using the
// real fixture files checked into the repository.
func TestParsePattern(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		wantName        string
		wantAuthor      string
		wantLicense     string
		wantFormatVer   int
		wantUserVer     int
		wantCategory    string
		wantTags        []string
		wantNotes       int
		wantInstTypes   []string
	}{
		{
			name:          "v2.0.0 modern namespaced",
			path:          "../../res/hydrogen-artifacts/v2.0.0.h2pattern",
			wantName:      "pat",
			wantAuthor:    "Hydrogen dev team",
			wantLicense:   "CC0",
			wantFormatVer: 2,
			wantUserVer:   0,
			wantCategory:  "unknown",
			wantTags:      []string{"Example", "Pattern"},
			wantNotes:     20,
			wantInstTypes: []string{"Hand Clap", "Kick", "Snare", "Stick"},
		},
		{
			name:          "v1.X.X namespaced legacy",
			path:          "../../res/hydrogen-artifacts/legacy-patterns/pattern-1.X.X.h2pattern",
			wantName:      "pat",
			wantAuthor:    "Hydrogen dev team",
			wantLicense:   "Public Domain",
			wantFormatVer: 0,
			wantUserVer:   0,
			wantCategory:  "unknown",
			wantTags:      []string{},
			wantNotes:     20,
			wantInstTypes: []string{},
		},
		{
			name:          "legacy no-namespace oldest",
			path:          "../../res/hydrogen-artifacts/legacy-patterns/legacy_pattern.h2pattern",
			wantName:      "Demo 1",
			wantAuthor:    "",
			wantLicense:   "",
			wantFormatVer: 0,
			wantUserVer:   0,
			wantCategory:  "demo-songs",
			wantTags:      []string{},
			wantNotes:     21, // fixture contains 21 notes
			wantInstTypes: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.path)
			if err != nil {
				t.Fatalf("open fixture: %v", err)
			}
			defer f.Close()

			got, err := parser.ParsePattern(f)
			if err != nil {
				t.Fatalf("ParsePattern: %v", err)
			}

			if got.Name != tc.wantName {
				t.Errorf("Name: got %q, want %q", got.Name, tc.wantName)
			}
			if got.Author != tc.wantAuthor {
				t.Errorf("Author: got %q, want %q", got.Author, tc.wantAuthor)
			}
			if got.License != tc.wantLicense {
				t.Errorf("License: got %q, want %q", got.License, tc.wantLicense)
			}
			if got.FormatVersion != tc.wantFormatVer {
				t.Errorf("FormatVersion: got %d, want %d", got.FormatVersion, tc.wantFormatVer)
			}
			if got.UserVersion != tc.wantUserVer {
				t.Errorf("UserVersion: got %d, want %d", got.UserVersion, tc.wantUserVer)
			}
			if got.Category != tc.wantCategory {
				t.Errorf("Category: got %q, want %q", got.Category, tc.wantCategory)
			}
			if got.Notes != tc.wantNotes {
				t.Errorf("Notes: got %d, want %d", got.Notes, tc.wantNotes)
			}

			// Tags: treat nil and empty slice as equivalent (no tags present).
			gotTags := got.Tags
			if gotTags == nil {
				gotTags = []string{}
			}
			if len(gotTags) != len(tc.wantTags) {
				t.Errorf("Tags: got %v (len %d), want %v (len %d)",
					gotTags, len(gotTags), tc.wantTags, len(tc.wantTags))
			} else {
				for i, tag := range tc.wantTags {
					if gotTags[i] != tag {
						t.Errorf("Tags[%d]: got %q, want %q", i, gotTags[i], tag)
					}
				}
			}

			// InstrumentTypes: treat nil and empty slice as equivalent.
			gotTypes := got.InstrumentTypes
			if gotTypes == nil {
				gotTypes = []string{}
			}
			if len(gotTypes) != len(tc.wantInstTypes) {
				t.Errorf("InstrumentTypes: got %v (len %d), want %v (len %d)",
					gotTypes, len(gotTypes), tc.wantInstTypes, len(tc.wantInstTypes))
			} else {
				for i, typ := range tc.wantInstTypes {
					if gotTypes[i] != typ {
						t.Errorf("InstrumentTypes[%d]: got %q, want %q", i, gotTypes[i], typ)
					}
				}
			}
		})
	}
}
