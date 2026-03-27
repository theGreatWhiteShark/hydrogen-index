// Package model defines the data types shared across hydrogen-index.
//
// These types represent both the parsed metadata from Hydrogen XML artifacts
// and the output index.json structure defined in MADR 0002.
package model

// Version is the current hydrogen-index version, used both for --version
// output and the "version" field in generated index.json files.
const Version = "0.1.0"

// ArtifactType identifies which category of Hydrogen artifact a file belongs to.
type ArtifactType string

const (
	ArtifactTypeDrumkit ArtifactType = "drumkit"
	ArtifactTypePattern ArtifactType = "pattern"
	ArtifactTypeSong    ArtifactType = "song"
)

// Index is the top-level structure of the generated index.json file.
// See MADR 0002 for the full specification.
type Index struct {
	Version      string           `json:"version"`
	Created      string           `json:"created"`
	PatternCount int              `json:"patternCount"`
	SongCount    int              `json:"songCount"`
	DrumkitCount int              `json:"drumkitCount"`
	Patterns     []PatternEntry   `json:"patterns"`
	Songs        []SongEntry      `json:"songs"`
	Drumkits     []DrumkitEntry   `json:"drumkits"`
	Hash         string           `json:"hash"`
}

// SharedFields contains the fields common to all artifact entries in the index.
// Embedded in PatternEntry, SongEntry, and DrumkitEntry.
type SharedFields struct {
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

// PatternEntry represents a single pattern artifact in the index.
type PatternEntry struct {
	SharedFields
	Notes           int      `json:"notes"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

// SongEntry represents a single song artifact in the index.
type SongEntry struct {
	SharedFields
	Patterns int `json:"patterns"`
}

// DrumkitEntry represents a single drumkit artifact in the index.
type DrumkitEntry struct {
	SharedFields
	Instruments     int      `json:"instruments"`
	Components      int      `json:"components"`
	Samples         int      `json:"samples"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

// DrumkitMetadata holds metadata extracted from a drumkit XML file.
// This intermediate type decouples XML parsing from index generation.
type DrumkitMetadata struct {
	Name            string
	Author          string
	Info            string
	License         string
	FormatVersion   int
	UserVersion     int
	Tags            []string
	Instruments     int
	Components      int
	Samples         int
	InstrumentTypes []string
}

// PatternMetadata holds metadata extracted from a pattern XML file.
type PatternMetadata struct {
	Name          string
	Author        string
	Info          string
	License       string
	FormatVersion int
	UserVersion   int
	Tags          []string
	Category      string
	Notes         int
	InstrumentTypes []string
}

// SongMetadata holds metadata extracted from a song XML file.
type SongMetadata struct {
	Name          string
	Author        string
	Info          string
	License       string
	FormatVersion int
	UserVersion   int
	Tags          []string
	Patterns      int
}
