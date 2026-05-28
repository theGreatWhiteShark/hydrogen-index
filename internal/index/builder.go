// Package index assembles index.json from scanner results.
package index

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"

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
// The hash is SHA-256 of the canonical compact JSON with the "hash" field
// removed. Canonical form uses alphabetically sorted keys at all nesting
// levels and preserves raw UTF-8 (no \uXXXX escaping), matching the C++
// parser in OnlineImporter.cpp which parses the JSON into QJsonObject and
// re-serializes with QJsonDocument::Compact.
// Returns the final JSON bytes (indented, with trailing newline) ready to
// write to disk.
func Finalize(idx *model.Index) ([]byte, error) {
	// Canonicalize: hash field must be empty before computing the digest so
	// that the hash is stable regardless of what the caller set it to.
	idx.Hash = ""

	// Marshal to compact JSON, then unmarshal into map[string]interface{} to
	// get canonical key ordering (Go sorts map keys alphabetically when
	// marshaling). This ensures the hash is independent of Go struct field
	// order and matches Qt's QJsonObject behavior.
	compact, err := json.Marshal(idx)
	if err != nil {
		return nil, fmt.Errorf("marshal for hashing: %w", err)
	}

	var canonical map[string]interface{}
	if err := json.Unmarshal(compact, &canonical); err != nil {
		return nil, fmt.Errorf("unmarshal for canonical form: %w", err)
	}

	// Remove the "hash" key to match C++ parser behavior
	// (QJsonObject::remove("hash") + QJsonDocument::Compact)
	delete(canonical, "hash")

	// Use canonicalJSON to produce Qt-compatible output: sorted keys + raw
	// UTF-8 (no \uXXXX escaping). Go's json.Marshal escapes <>& and all
	// non-ASCII as \uXXXX, but QJsonDocument::toJson preserves raw UTF-8.
	// Using a json.Encoder with SetEscapeHTML(false) and a custom string
	// encoder ensures byte-for-byte compatibility.
	dataWithoutHash, err := canonicalJSON(canonical)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical: %w", err)
	}

	digest := sha256.Sum256(dataWithoutHash)
	idx.Hash = hex.EncodeToString(digest[:])

	final, err := marshalIndented(idx)
	if err != nil {
		return nil, fmt.Errorf("marshal final: %w", err)
	}

	return final, nil
}

// canonicalJSON produces compact JSON with sorted keys at all nesting levels
// and raw UTF-8 encoding (no \uXXXX escaping). This matches the output of
// Qt's QJsonDocument::toJson(QJsonDocument::Compact) when the source
// QJsonObject has alphabetically ordered keys.
func canonicalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := canonicalMarshal(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// canonicalMarshal writes v to w as compact JSON with sorted keys and raw
// UTF-8. It mirrors json.Marshal's output format but avoids Go's default
// HTML-safe and ASCII-safe escaping, ensuring compatibility with Qt's
// QJsonDocument serialization.
func canonicalMarshal(w *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case nil:
		w.WriteString("null")
	case bool:
		w.WriteString(strconv.FormatBool(val))
	case int:
		w.WriteString(strconv.FormatInt(int64(val), 10))
	case int64:
		w.WriteString(strconv.FormatInt(val, 10))
	case float64:
		// Use Go's json.Number format for floats to match json.Marshal
		w.WriteString(json.Number(strconv.FormatFloat(val, 'f', -1, 64)).String())
	case string:
		writeJSONString(w, val)
	case []any:
		w.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				w.WriteByte(',')
			}
			if err := canonicalMarshal(w, item); err != nil {
				return err
			}
		}
		w.WriteByte(']')
	case map[string]any:
		w.WriteByte('{')
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				w.WriteByte(',')
			}
			writeJSONString(w, k)
			w.WriteByte(':')
			if err := canonicalMarshal(w, val[k]); err != nil {
				return err
			}
		}
		w.WriteByte('}')
	default:
		return fmt.Errorf("unsupported type %T in canonicalJSON", v)
	}
	return nil
}

// writeJSONString writes a JSON-encoded string to w with raw UTF-8
// (no \uXXXX escaping of non-ASCII). This matches Qt's QJsonStringEncoder
// behavior which only escapes control chars, quotes, and backslashes.
func writeJSONString(w *bytes.Buffer, s string) {
	w.WriteByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch r {
		case '"':
			w.WriteString(`\"`)
		case '\\':
			w.WriteString(`\\`)
		case '\b':
			w.WriteString(`\b`)
		case '\f':
			w.WriteString(`\f`)
		case '\n':
			w.WriteString(`\n`)
		case '\r':
			w.WriteString(`\r`)
		case '\t':
			w.WriteString(`\t`)
		default:
			if r < 0x20 {
				// Control characters: \u00XX
				fmt.Fprintf(w, "\\u%04x", r)
			} else {
				// All other characters (including non-ASCII): write raw UTF-8
				w.WriteString(s[i : i+size])
			}
		}
		i += size
	}
	w.WriteByte('"')
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
	url := a.RelPath
	if a.BaseURL != "" {
		url = a.BaseURL + "/" + a.RelPath
	}
	return model.SharedFields{
		Type:          artifactType,
		Name:          name,
		URL:           url,
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
		FolderName:      m.FolderName,
		Instruments:     m.Instruments,
		Components:      m.Components,
		Samples:         m.Samples,
		InstrumentTypes: instrumentTypes,
	}
}
