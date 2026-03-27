package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/hydrogen-music/hydrogen-index/pkg/domain"
)

type h2Tags struct {
	Tag []string `xml:"tag"`
}

type commonFields struct {
	Name          string `xml:"name"`
	Author        string `xml:"author"`
	Info          string `xml:"info"`
	UserVersion   int    `xml:"userVersion"`
	FormatVersion int    `xml:"formatVersion"`
	Tags          h2Tags `xml:"tags"`
	License       string `xml:"license"`
}

type patternXML struct {
	XMLName xml.Name `xml:"drumkit_pattern"`
	Pattern struct {
		commonFields
		NoteList struct {
			Notes []struct {
				Type string `xml:"type"`
			} `xml:"note"`
		} `xml:"noteList"`
	} `xml:"pattern"`
}

type songXML struct {
	XMLName     xml.Name `xml:"song"`
	commonFields
	PatternList struct {
		Patterns []interface{} `xml:"pattern"`
	} `xml:"patternList"`
}

type drumkitXML struct {
	XMLName xml.Name `xml:"drumkit_info"`
	commonFields
	InstrumentList struct {
		Instruments []struct {
			Type                string `xml:"type"`
			InstrumentComponent []struct {
				Layer []struct {
					Filename string `xml:"filename"`
				} `xml:"layer"`
			} `xml:"instrumentComponent"`
		} `xml:"instrument"`
	} `xml:"instrumentList"`
}

func computeHashAndSize(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", 0, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(h.Sum(nil)), stat.Size(), nil
}

func deduplicate(input []string) []string {
	m := make(map[string]bool)
	var result []string
	for _, s := range input {
		if s != "" && !m[s] {
			m[s] = true
			result = append(result, s)
		}
	}
	sort.Strings(result)
	return result
}

func ParseArtifact(xmlPath string, originalPath string) (interface{}, error) {
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	hash, size, err := computeHashAndSize(originalPath)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(xmlFile)
	// We peek at the root element to decide what to parse
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	for {
		if start, ok := token.(xml.StartElement); ok {
			switch start.Name.Local {
			case "drumkit_pattern":
				var p patternXML
				if err := decoder.DecodeElement(&p, &start); err != nil {
					return nil, err
				}
				var types []string
				for _, n := range p.Pattern.NoteList.Notes {
					types = append(types, n.Type)
				}
				return &domain.PatternBlock{
					CommonBlock: domain.CommonBlock{
						Type:          "pattern",
						Name:          p.Pattern.Name,
						Hash:          hash,
						Size:          size,
						Author:        p.Pattern.Author,
						Description:   p.Pattern.Info,
						Version:       p.Pattern.UserVersion,
						FormatVersion: p.Pattern.FormatVersion,
						Tags:          p.Pattern.Tags.Tag,
						License:       p.Pattern.License,
					},
					Notes:           len(p.Pattern.NoteList.Notes),
					InstrumentTypes: deduplicate(types),
				}, nil
			case "song":
				var s songXML
				if err := decoder.DecodeElement(&s, &start); err != nil {
					return nil, err
				}
				return &domain.SongBlock{
					CommonBlock: domain.CommonBlock{
						Type:          "song",
						Name:          s.Name,
						Hash:          hash,
						Size:          size,
						Author:        s.Author,
						Description:   s.Info,
						Version:       s.UserVersion,
						FormatVersion: s.FormatVersion,
						Tags:          s.Tags.Tag,
						License:       s.License,
					},
					Patterns: len(s.PatternList.Patterns),
				}, nil
			case "drumkit_info":
				var d drumkitXML
				if err := decoder.DecodeElement(&d, &start); err != nil {
					return nil, err
				}
				var types []string
				components := 0
				samples := 0
				for _, inst := range d.InstrumentList.Instruments {
					types = append(types, inst.Type)
					components += len(inst.InstrumentComponent)
					for _, comp := range inst.InstrumentComponent {
						samples += len(comp.Layer)
					}
				}
				return &domain.DrumkitBlock{
					CommonBlock: domain.CommonBlock{
						Type:          "drumkit",
						Name:          d.Name,
						Hash:          hash,
						Size:          size,
						Author:        d.Author,
						Description:   d.Info,
						Version:       d.UserVersion,
						FormatVersion: d.FormatVersion,
						Tags:          d.Tags.Tag,
						License:       d.License,
					},
					Instruments:     len(d.InstrumentList.Instruments),
					Components:      components,
					Samples:         samples,
					InstrumentTypes: deduplicate(types),
				}, nil
			}
		}
		token, err = decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return nil, fmt.Errorf("unknown or invalid hydrogen artifact")
}
