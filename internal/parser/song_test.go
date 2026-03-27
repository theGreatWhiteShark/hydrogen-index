package parser

import (
	"io"
	"os"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
)

func TestParseSong(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected model.SongMetadata
	}{
		{
			name: "v2.0.0 modern format",
			file: "../../res/hydrogen-artifacts/v2.0.0.h2song",
			expected: model.SongMetadata{
				Name:          "Untitled Song",
				Author:        "hydrogen",
				Info:          "...",
				License:       "CC0",
				FormatVersion: 2,
				UserVersion:   0,
				Tags:          []string{"Example", "Song"},
				Patterns:      10,
			},
		},
		{
			name: "v1.2.2 no formatVersion or tags",
			file: "../../res/hydrogen-artifacts/legacy-songs/test_song_1.2.2.h2song",
			expected: model.SongMetadata{
				Name:          "Untitled Song",
				Author:        "hydrogen",
				Info:          "...",
				License:       "undefined license",
				FormatVersion: 0,
				UserVersion:   0,
				Tags:          []string{},
				Patterns:      10,
			},
		},
		{
			name: "v0.9.3 oldest format, no license",
			file: "../../res/hydrogen-artifacts/legacy-songs/test_song_0.9.3.h2song",
			expected: model.SongMetadata{
				Name:          "Untitled Song",
				Author:        "kzapfe",
				Info:          "Licensed under GPLv2+",
				License:       "",
				FormatVersion: 0,
				UserVersion:   0,
				Tags:          []string{},
				Patterns:      10,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.file)
			if err != nil {
				t.Fatalf("open fixture: %v", err)
			}
			defer f.Close()

			got, err := ParseSong(f)
			if err != nil {
				t.Fatalf("ParseSong error: %v", err)
			}

			if got.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", got.Name, tc.expected.Name)
			}
			if got.Author != tc.expected.Author {
				t.Errorf("Author: got %q, want %q", got.Author, tc.expected.Author)
			}
			if got.Info != tc.expected.Info {
				t.Errorf("Info: got %q, want %q", got.Info, tc.expected.Info)
			}
			if got.License != tc.expected.License {
				t.Errorf("License: got %q, want %q", got.License, tc.expected.License)
			}
			if got.FormatVersion != tc.expected.FormatVersion {
				t.Errorf("FormatVersion: got %d, want %d", got.FormatVersion, tc.expected.FormatVersion)
			}
			if got.UserVersion != tc.expected.UserVersion {
				t.Errorf("UserVersion: got %d, want %d", got.UserVersion, tc.expected.UserVersion)
			}
			if got.Patterns != tc.expected.Patterns {
				t.Errorf("Patterns: got %d, want %d", got.Patterns, tc.expected.Patterns)
			}

			// Tags: compare length then each element.
			if len(got.Tags) != len(tc.expected.Tags) {
				t.Errorf("Tags length: got %d, want %d (got %v, want %v)",
					len(got.Tags), len(tc.expected.Tags), got.Tags, tc.expected.Tags)
			} else {
				for i, tag := range tc.expected.Tags {
					if got.Tags[i] != tag {
						t.Errorf("Tags[%d]: got %q, want %q", i, got.Tags[i], tag)
					}
				}
			}
		})
	}
}

func TestParseSongMalformedXML(t *testing.T) {
	// Malformed XML must return an error, not silently succeed.
	_, err := ParseSong(errReader("not xml at all <<<<"))
	if err == nil {
		t.Error("expected error for malformed XML, got nil")
	}
}

// errReader is a helper that wraps a string as an io.Reader for small test inputs.
type errReader string

func (s errReader) Read(p []byte) (int, error) {
	n := copy(p, []byte(s))
	return n, io.EOF
}
