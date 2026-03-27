package validate

import (
	"testing"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
)

func TestValidateBytesAcceptsValidDocument(t *testing.T) {
	t.Parallel()

	data, err := indexfile.Marshal(indexfile.Document{
		Version:      "0.1.0",
		Created:      "2026-03-27T14:00:00",
		PatternCount: 1,
		SongCount:    1,
		DrumkitCount: 1,
		Patterns: []indexfile.PatternArtifact{{
			Artifact: indexfile.Artifact{
				Type:          indexfile.ArtifactTypePattern,
				Name:          "Pattern",
				URL:           "https://example.com/pattern.h2pattern",
				Hash:          "3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986",
				Author:        "Name",
				Description:   "Description",
				Version:       1,
				FormatVersion: 2,
				Tags:          []string{"Example"},
				Size:          123,
				License:       "CC0",
			},
			Notes:           5,
			InstrumentTypes: []string{"Kick"},
		}},
		Songs: []indexfile.SongArtifact{{
			Artifact: indexfile.Artifact{
				Type:          indexfile.ArtifactTypeSong,
				Name:          "Song",
				URL:           "https://example.com/song.h2song",
				Hash:          "3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986",
				Author:        "Name",
				Description:   "Description",
				Version:       1,
				FormatVersion: 2,
				Tags:          []string{},
				Size:          456,
				License:       "CC0",
			},
			Patterns: 3,
		}},
		Drumkits: []indexfile.DrumkitArtifact{{
			Artifact: indexfile.Artifact{
				Type:          indexfile.ArtifactTypeDrumkit,
				Name:          "Kit",
				URL:           "https://example.com/kit.h2drumkit",
				Hash:          "3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986",
				Author:        "Name",
				Description:   "Description",
				Version:       1,
				FormatVersion: 2,
				Tags:          []string{"Example"},
				Size:          789,
				License:       "CC0",
			},
			Instruments:     2,
			Components:      2,
			Samples:         4,
			InstrumentTypes: []string{"Kick", "Snare"},
		}},
		Hash: "776aa3fd387d517b6799181389efe2e9baf51fe03e2ede17c9488daf3b13883d",
	})
	if err != nil {
		t.Fatalf("indexfile.Marshal() returned error: %v", err)
	}

	if err := ValidateBytes(data); err != nil {
		t.Fatalf("ValidateBytes() returned error: %v", err)
	}
}

func TestValidateBytesRejectsMalformedDocument(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"version": "0.1.0",
		"created": "2026-03-27T14:00:00",
		"patternCount": "oops",
		"songCount": 0,
		"drumkitCount": 0,
		"patterns": [],
		"songs": [],
		"drumkits": []
	}`)

	if err := ValidateBytes(data); err == nil {
		t.Fatal("ValidateBytes() unexpectedly accepted malformed input")
	}
}
