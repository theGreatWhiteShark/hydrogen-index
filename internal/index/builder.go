// Package index assembles index.json from scanner results.
package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
	"github.com/hydrogen-music/hydrogen-index/internal/scanner"
)

// Build constructs an Index from a slice of discovered artifacts.
// The returned Index has all fields populated except Hash (computed by Finalize).
func Build(artifacts []scanner.ArtifactFile) (*model.Index, error) {
	idx := &model.Index{
		Version:  model.Version,
		Created:  time.Now().UTC().Format("2006-01-02T15:04:05"),
		Patterns: []model.PatternEntry{},
		Songs:    []model.SongEntry{},
		Drumkits: []model.DrumkitEntry{},
	}

	for _, a := range artifacts {
		switch m := a.Metadata.(type) {
		case *model.PatternMetadata:
			idx.Patterns = append(idx.Patterns, patternEntry(a, m))
		case *model.SongMetadata:
			idx.Songs = append(idx.Songs, songEntry(a, m))
		case *model.DrumkitMetadata:
			idx.Drumkits = append(idx.Drumkits, drumkitEntry(a, m))
		default:
			return nil, fmt.Errorf("unknown metadata type %T for artifact %s", a.Metadata, a.RelPath)
		}
	}

	idx.PatternCount = len(idx.Patterns)
	idx.SongCount = len(idx.Songs)
	idx.DrumkitCount = len(idx.Drumkits)

	return idx, nil
}

// Finalize serializes the Index to JSON and computes the self-hash.
// The hash is SHA-256 of the JSON with the "hash" field set to "".
// Returns the final JSON bytes (with trailing newline) ready to write to disk.
func Finalize(idx *model.Index) ([]byte, error) {
	// Canonicalize: hash field must be empty before computing the digest so
	// that the hash is stable regardless of what the caller set it to.
	idx.Hash = ""

	preliminary, err := marshalIndented(idx)
	if err != nil {
		return nil, fmt.Errorf("marshal for hashing: %w", err)
	}

	digest := sha256.Sum256(preliminary)
	idx.Hash = hex.EncodeToString(digest[:])

	final, err := marshalIndented(idx)
	if err != nil {
		return nil, fmt.Errorf("marshal final: %w", err)
	}

	return final, nil
}

// marshalIndented encodes v as indented JSON with a trailing newline.
func marshalIndented(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// sharedFields maps the common artifact metadata to SharedFields.
func sharedFields(a scanner.ArtifactFile, artifactType model.ArtifactType,
	name, author, info, license string,
	userVersion, formatVersion int,
	tags []string,
) model.SharedFields {
	if tags == nil {
		tags = []string{}
	}
	return model.SharedFields{
		Type:          artifactType,
		Name:          name,
		URL:           a.RelPath,
		Hash:          a.Hash,
		Author:        author,
		Description:   info,
		Version:       userVersion,
		FormatVersion: formatVersion,
		Tags:          tags,
		Size:          a.Size,
		License:       license,
	}
}

func patternEntry(a scanner.ArtifactFile, m *model.PatternMetadata) model.PatternEntry {
	instrumentTypes := m.InstrumentTypes
	if instrumentTypes == nil {
		instrumentTypes = []string{}
	}
	return model.PatternEntry{
		SharedFields:    sharedFields(a, model.ArtifactTypePattern, m.Name, m.Author, m.Info, m.License, m.UserVersion, m.FormatVersion, m.Tags),
		Notes:           m.Notes,
		InstrumentTypes: instrumentTypes,
	}
}

func songEntry(a scanner.ArtifactFile, m *model.SongMetadata) model.SongEntry {
	return model.SongEntry{
		SharedFields: sharedFields(a, model.ArtifactTypeSong, m.Name, m.Author, m.Info, m.License, m.UserVersion, m.FormatVersion, m.Tags),
		Patterns:     m.Patterns,
	}
}

func drumkitEntry(a scanner.ArtifactFile, m *model.DrumkitMetadata) model.DrumkitEntry {
	instrumentTypes := m.InstrumentTypes
	if instrumentTypes == nil {
		instrumentTypes = []string{}
	}
	return model.DrumkitEntry{
		SharedFields:    sharedFields(a, model.ArtifactTypeDrumkit, m.Name, m.Author, m.Info, m.License, m.UserVersion, m.FormatVersion, m.Tags),
		Instruments:     m.Instruments,
		Components:      m.Components,
		Samples:         m.Samples,
		InstrumentTypes: instrumentTypes,
	}
}
