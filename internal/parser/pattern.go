// Package parser implements XML parsers for Hydrogen artifact file formats.
package parser

import (
	"encoding/xml"
	"io"
	"sort"
	"strings"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
)

// xmlNote holds the fields we care about from a <note> element across all format
// versions. The <type> field only appears in v2.0.0+.
type xmlNote struct {
	Type string `xml:"type"`
}

// xmlPattern holds the fields inside the <pattern> element.
// Both <name> (modern) and <pattern_name> (legacy) are captured so we can fall
// back to the legacy spelling.
type xmlPattern struct {
	FormatVersion int       `xml:"formatVersion"`
	UserVersion   int       `xml:"userVersion"`
	Name          string    `xml:"name"`
	PatternName   string    `xml:"pattern_name"` // legacy alias for name
	Author        string    `xml:"author"`
	Info          string    `xml:"info"`
	License       string    `xml:"license"`
	Category      string    `xml:"category"`
	Tags          []string  `xml:"tags>tag"`
	Notes         []xmlNote `xml:"noteList>note"`
}

// xmlRoot captures both namespaced and non-namespaced <drumkit_pattern> roots.
// Author and License appear here in the v1.X.X format (not inside <pattern>).
type xmlRoot struct {
	// Author and License at root level (v1.X.X format only).
	Author  string     `xml:"author"`
	License string     `xml:"license"`
	Pattern xmlPattern `xml:"pattern"`
}

// ParsePattern parses a pattern XML file (.h2pattern) and returns metadata.
//
// Three format variants are handled:
//   - v2.0.0: namespaced root, author/license inside <pattern>, typed notes
//   - v1.X.X: namespaced root, author/license at root level, no types
//   - legacy:  no namespace, <pattern_name> instead of <name>, no author/license
func ParsePattern(r io.Reader) (*model.PatternMetadata, error) {
	// Decode the XML into a namespace-agnostic struct. Go's encoding/xml matches
	// on local element names, so the namespace attribute on the root element is
	// silently ignored and all three formats decode into the same struct layout.
	var root xmlRoot
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	pat := root.Pattern

	// Resolve Name: modern formats use <name>; the oldest legacy format uses
	// <pattern_name>. Prefer <name> if non-empty.
	name := strings.TrimSpace(pat.Name)
	if name == "" {
		name = strings.TrimSpace(pat.PatternName)
	}

	// Resolve Author/License: v2.0.0 stores them inside <pattern>; v1.X.X stores
	// them at the <drumkit_pattern> root level; legacy has neither.
	author := strings.TrimSpace(pat.Author)
	if author == "" {
		author = strings.TrimSpace(root.Author)
	}

	license := strings.TrimSpace(pat.License)
	if license == "" {
		license = strings.TrimSpace(root.License)
	}

	// Collect instrument types: sorted, deduplicated, non-empty strings from
	// <type> child elements of each <note>. Only v2.0.0+ has <type>; earlier
	// formats omit it, so this naturally produces an empty slice for them.
	typeSet := make(map[string]struct{})
	for _, n := range pat.Notes {
		t := strings.TrimSpace(n.Type)
		if t != "" {
			typeSet[t] = struct{}{}
		}
	}
	instrumentTypes := make([]string, 0, len(typeSet))
	for t := range typeSet {
		instrumentTypes = append(instrumentTypes, t)
	}
	sort.Strings(instrumentTypes)

	return &model.PatternMetadata{
		Name:            name,
		Author:          author,
		Info:            strings.TrimSpace(pat.Info),
		License:         license,
		FormatVersion:   pat.FormatVersion,
		UserVersion:     pat.UserVersion,
		Tags:            pat.Tags,
		Category:        strings.TrimSpace(pat.Category),
		Notes:           len(pat.Notes),
		InstrumentTypes: instrumentTypes,
	}, nil
}
