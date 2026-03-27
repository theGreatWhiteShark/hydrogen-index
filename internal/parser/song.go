package parser

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
)

// xmlSongRoot captures the fields we need from a <song> element.
// Songs have no namespace — the root is always <song>.
type xmlSongRoot struct {
	FormatVersion int      `xml:"formatVersion"`
	UserVersion   int      `xml:"userVersion"`
	Name          string   `xml:"name"`
	Author        string   `xml:"author"`
	Notes         string   `xml:"notes"`
	License       string   `xml:"license"`
	Tags          []string `xml:"tags>tag"`
}

// ParseSong parses a song XML file (.h2song) and returns metadata.
//
// Three format variants are handled:
//   - v2.0.0: has formatVersion, userVersion, tags
//   - v1.2.2: no formatVersion/userVersion/tags, has license
//   - v0.9.3: oldest; no formatVersion/userVersion/tags/license
//
// Pattern count is determined by a second token-based pass because
// we must count only <pattern> elements that are direct children of
// the top-level <patternList>, not any pattern-like structures nested
// inside <instrumentList> or other sections.
func ParseSong(r io.Reader) (*model.SongMetadata, error) {
	// Buffer the entire stream so we can decode it twice — once for the struct
	// fields and once for the streaming pattern count.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var root xmlSongRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	patterns, err := countSongPatterns(data)
	if err != nil {
		return nil, err
	}

	// Tags missing from older formats should be an empty (non-nil) slice.
	tags := root.Tags
	if tags == nil {
		tags = []string{}
	}

	return &model.SongMetadata{
		Name:          strings.TrimSpace(root.Name),
		Author:        strings.TrimSpace(root.Author),
		Info:          strings.TrimSpace(root.Notes),
		License:       strings.TrimSpace(root.License),
		FormatVersion: root.FormatVersion,
		UserVersion:   root.UserVersion,
		Tags:          tags,
		Patterns:      patterns,
	}, nil
}

// countSongPatterns counts <pattern> elements that are direct children of the
// top-level <patternList> element (which is itself a direct child of <song>).
//
// We use a token-based walk rather than struct decoding because the struct
// approach cannot cheaply distinguish the top-level <patternList> from any
// nested pattern-like structures that may appear inside <instrumentList>.
// We track depth relative to each section so we can be precise about which
// <pattern> start tags we tally.
func countSongPatterns(data []byte) (int, error) {
	dec := xml.NewDecoder(strings.NewReader(string(data)))

	// depth tracks how many XML elements we are currently inside.
	depth := 0
	// patternListDepth is set to the depth at which we entered <patternList>
	// as a child of <song> (depth==1). Zero means we are not inside it.
	patternListDepth := 0
	count := 0

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch {
			// Enter the top-level <patternList> (direct child of <song>, depth==2).
			case t.Name.Local == "patternList" && depth == 2:
				patternListDepth = depth

			// Count <pattern> elements that are direct children of <patternList>
			// (depth == patternListDepth+1).
			case patternListDepth != 0 &&
				t.Name.Local == "pattern" &&
				depth == patternListDepth+1:
				count++
			}

		case xml.EndElement:
			// Leave <patternList>: stop counting.
			if patternListDepth != 0 && depth == patternListDepth {
				patternListDepth = 0
			}
			depth--
		}
	}

	return count, nil
}
