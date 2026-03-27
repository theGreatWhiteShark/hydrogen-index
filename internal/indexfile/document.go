package indexfile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type ArtifactType string

const (
	ArtifactTypePattern ArtifactType = "pattern"
	ArtifactTypeSong    ArtifactType = "song"
	ArtifactTypeDrumkit ArtifactType = "drumkit"
)

type Artifact struct {
	Type          ArtifactType `json:"type"`
	Name          string       `json:"name"`
	URL           string       `json:"url"`
	Hash          string       `json:"hash"`
	Author        string       `json:"author"`
	Description   string       `json:"description"`
	Version       int          `json:"version"`
	FormatVersion int          `json:"formatVersion"`
	Tags          []string     `json:"tags"`
	Size          int64        `json:"size"`
	License       string       `json:"license"`
}

type PatternArtifact struct {
	Artifact
	Notes           int      `json:"notes"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

type SongArtifact struct {
	Artifact
	Patterns int `json:"patterns"`
}

type DrumkitArtifact struct {
	Artifact
	Instruments     int      `json:"instruments"`
	Components      int      `json:"components"`
	Samples         int      `json:"samples"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

type Document struct {
	Version      string            `json:"version"`
	Created      string            `json:"created"`
	PatternCount int               `json:"patternCount"`
	SongCount    int               `json:"songCount"`
	DrumkitCount int               `json:"drumkitCount"`
	Patterns     []PatternArtifact `json:"patterns"`
	Songs        []SongArtifact    `json:"songs"`
	Drumkits     []DrumkitArtifact `json:"drumkits"`
	Hash         string            `json:"hash,omitempty"`
}

func Marshal(document Document) ([]byte, error) {
	document.Patterns = ensurePatternSlice(document.Patterns)
	document.Songs = ensureSongSlice(document.Songs)
	document.Drumkits = ensureDrumkitSlice(document.Drumkits)

	withoutHash := document
	withoutHash.Hash = ""

	body, err := json.MarshalIndent(withoutHash, "", "  ")
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(body)
	document.Hash = hex.EncodeToString(sum[:])

	return json.MarshalIndent(document, "", "  ")
}

func ensurePatternSlice(values []PatternArtifact) []PatternArtifact {
	if values == nil {
		return []PatternArtifact{}
	}
	return values
}

func ensureSongSlice(values []SongArtifact) []SongArtifact {
	if values == nil {
		return []SongArtifact{}
	}
	return values
}

func ensureDrumkitSlice(values []DrumkitArtifact) []DrumkitArtifact {
	if values == nil {
		return []DrumkitArtifact{}
	}
	return values
}
