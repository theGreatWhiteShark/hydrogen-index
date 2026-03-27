package domain

import (
	"encoding/json"
)

// CommonBlock contains the shared properties of all artifact blocks in the index file.
type CommonBlock struct {
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Hash          string   `json:"hash"`
	Author        string   `json:"author"`
	Description   string   `json:"description"`
	Version       int      `json:"version"`
	FormatVersion int      `json:"formatVersion"`
	Tags          []string `json:"tags"`
	Size          int64    `json:"size"`
	License       string   `json:"license"`
}

// PatternBlock represents a single Hydrogen pattern entry in the index file.
type PatternBlock struct {
	CommonBlock
	Notes           int      `json:"notes"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

func (p *PatternBlock) UnmarshalJSON(data []byte) error {
	type Alias PatternBlock
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return json.Unmarshal(data, &p.CommonBlock)
}

// SongBlock represents a single Hydrogen song entry in the index file.
type SongBlock struct {
	CommonBlock
	Patterns int `json:"patterns"`
}

func (s *SongBlock) UnmarshalJSON(data []byte) error {
	type Alias SongBlock
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return json.Unmarshal(data, &s.CommonBlock)
}

// DrumkitBlock represents a single Hydrogen drumkit entry in the index file.
type DrumkitBlock struct {
	CommonBlock
	Instruments     int      `json:"instruments"`
	Components      int      `json:"components"`
	Samples         int      `json:"samples"`
	InstrumentTypes []string `json:"instrumentTypes"`
}

func (d *DrumkitBlock) UnmarshalJSON(data []byte) error {
	type Alias DrumkitBlock
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return json.Unmarshal(data, &d.CommonBlock)
}

// IndexFile is the top-level structure of the Hydrogen index file.
type IndexFile struct {
	Version      string         `json:"version"`
	Created      string         `json:"created"`
	PatternCount int            `json:"patternCount"`
	SongCount    int            `json:"songCount"`
	DrumkitCount int            `json:"drumkitCount"`
	Patterns     []PatternBlock `json:"patterns"`
	Songs        []SongBlock    `json:"songs"`
	Drumkits     []DrumkitBlock `json:"drumkits"`
	Hash         string         `json:"hash,omitempty"`
}
