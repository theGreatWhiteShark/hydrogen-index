// Package parser provides parsers for Hydrogen XML artifact formats.
package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/hydrogen-music/hydrogen-index/internal/model"
)

// ParseDrumkit parses a drumkit XML file (drumkit.xml content) and returns metadata.
// It handles all known Hydrogen drumkit format variants.
func ParseDrumkit(r io.Reader) (*model.DrumkitMetadata, error) {
	raw, err := parseRawDrumkit(r)
	if err != nil {
		return nil, err
	}
	return raw.toMetadata(), nil
}

// rawDrumkit is the intermediate parsed representation before interpretation.
// It captures the superset of fields across all format variants.
type rawDrumkit struct {
	name          string
	author        string
	info          string
	license       string
	formatVersion int
	userVersion   int
	tags          []string
	// top-level component list (v1.2.3 format)
	componentList []rawComponent
	instruments   []rawInstrument
}

type rawComponent struct {
	id   int
	name string
}

type rawInstrument struct {
	instrumentType string
	// instrComponents is non-nil when <instrumentComponent> wrappers are present
	instrComponents []rawInstrComponent
	// directLayers is used when layers live directly under <instrument>
	directLayers int
}

type rawInstrComponent struct {
	// name is used in v2.0.0; component_id is used in v1.2.3
	name        string
	componentID int
	layers      int
}

// parseRawDrumkit drives the XML decoder, dispatching on local element names
// to remain namespace-agnostic across all format variants.
func parseRawDrumkit(r io.Reader) (*rawDrumkit, error) {
	dec := xml.NewDecoder(r)

	// Advance to the root <drumkit_info> element.
	if err := seekElement(dec, "drumkit_info"); err != nil {
		return nil, fmt.Errorf("drumkit XML missing root element: %w", err)
	}

	raw := &rawDrumkit{}
	if err := parseDrumkitInfo(dec, raw); err != nil {
		return nil, fmt.Errorf("parsing drumkit_info: %w", err)
	}
	return raw, nil
}

// parseDrumkitInfo reads the children of <drumkit_info>.
func parseDrumkitInfo(dec *xml.Decoder, raw *rawDrumkit) error {
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			switch local {
			case "name":
				raw.name, err = charData(dec, local)
			case "author":
				raw.author, err = charData(dec, local)
			case "info":
				raw.info, err = charData(dec, local)
			case "license":
				raw.license, err = charData(dec, local)
			case "formatVersion":
				var v int
				v, err = intCharData(dec, local)
				if err == nil {
					raw.formatVersion = v
				}
			case "userVersion":
				var v int
				v, err = intCharData(dec, local)
				if err == nil {
					raw.userVersion = v
				}
			case "tags":
				raw.tags, err = parseTags(dec)
			case "componentList":
				raw.componentList, err = parseComponentList(dec)
			case "instrumentList":
				raw.instruments, err = parseInstrumentList(dec)
			default:
				err = skipElement(dec, local)
			}
			if err != nil {
				return fmt.Errorf("element <%s>: %w", local, err)
			}

		case xml.EndElement:
			// End of drumkit_info.
			return nil
		}
	}
}

// parseTags reads <tags><tag>...</tag></tags>.
func parseTags(dec *xml.Decoder) ([]string, error) {
	var tags []string
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			if local == "tag" {
				val, err := charData(dec, local)
				if err != nil {
					return nil, err
				}
				tags = append(tags, val)
			} else {
				if err := skipElement(dec, local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			return tags, nil
		}
	}
}

// parseComponentList reads <componentList><drumkitComponent>...</drumkitComponent></componentList>.
func parseComponentList(dec *xml.Decoder) ([]rawComponent, error) {
	var components []rawComponent
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			if local == "drumkitComponent" {
				comp, err := parseDrumkitComponent(dec)
				if err != nil {
					return nil, err
				}
				components = append(components, comp)
			} else {
				if err := skipElement(dec, local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			return components, nil
		}
	}
}

// parseDrumkitComponent reads the children of a single <drumkitComponent>.
func parseDrumkitComponent(dec *xml.Decoder) (rawComponent, error) {
	var comp rawComponent
	for {
		tok, err := dec.Token()
		if err != nil {
			return comp, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			switch local {
			case "id":
				comp.id, err = intCharData(dec, local)
			case "name":
				comp.name, err = charData(dec, local)
			default:
				err = skipElement(dec, local)
			}
			if err != nil {
				return comp, fmt.Errorf("drumkitComponent <%s>: %w", local, err)
			}
		case xml.EndElement:
			return comp, nil
		}
	}
}

// parseInstrumentList reads <instrumentList><instrument>...</instrument></instrumentList>.
func parseInstrumentList(dec *xml.Decoder) ([]rawInstrument, error) {
	var instruments []rawInstrument
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			if local == "instrument" {
				instr, err := parseInstrument(dec)
				if err != nil {
					return nil, err
				}
				instruments = append(instruments, instr)
			} else {
				if err := skipElement(dec, local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			return instruments, nil
		}
	}
}

// parseInstrument reads the children of a single <instrument>.
// It must distinguish between the three structural patterns:
//   - v2.0.0: <instrumentComponent name="Main"><layer>
//   - v1.2.3: <instrumentComponent component_id="N"><layer>
//   - legacy:  <layer> directly under <instrument>
func parseInstrument(dec *xml.Decoder) (rawInstrument, error) {
	var instr rawInstrument
	for {
		tok, err := dec.Token()
		if err != nil {
			return instr, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			switch local {
			case "type":
				instr.instrumentType, err = charData(dec, local)
			case "layer":
				// Direct layer under <instrument> — legacy format.
				instr.directLayers++
				err = skipElement(dec, local)
			case "instrumentComponent":
				comp, err := parseInstrComponent(dec)
				if err != nil {
					return instr, fmt.Errorf("instrumentComponent: %w", err)
				}
				instr.instrComponents = append(instr.instrComponents, comp)
			default:
				err = skipElement(dec, local)
			}
			if err != nil {
				return instr, fmt.Errorf("instrument <%s>: %w", local, err)
			}
		case xml.EndElement:
			return instr, nil
		}
	}
}

// parseInstrComponent reads a single <instrumentComponent>.
func parseInstrComponent(dec *xml.Decoder) (rawInstrComponent, error) {
	var comp rawInstrComponent
	for {
		tok, err := dec.Token()
		if err != nil {
			return comp, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t)
			switch local {
			case "name":
				comp.name, err = charData(dec, local)
			case "component_id":
				comp.componentID, err = intCharData(dec, local)
			case "layer":
				comp.layers++
				err = skipElement(dec, local)
			default:
				err = skipElement(dec, local)
			}
			if err != nil {
				return comp, fmt.Errorf("instrumentComponent <%s>: %w", local, err)
			}
		case xml.EndElement:
			return comp, nil
		}
	}
}

// toMetadata converts the raw parsed data into the public DrumkitMetadata type,
// applying the component-counting rules that depend on the format variant.
func (raw *rawDrumkit) toMetadata() *model.DrumkitMetadata {
	meta := &model.DrumkitMetadata{
		Name:          raw.name,
		Author:        raw.author,
		Info:          raw.info,
		License:       raw.license,
		FormatVersion: raw.formatVersion,
		UserVersion:   raw.userVersion,
		Tags:          nilToEmpty(raw.tags),
	}

	meta.Instruments = len(raw.instruments)

	components, samples := raw.countComponentsAndSamples()
	meta.Components = components
	meta.Samples = samples

	meta.InstrumentTypes = raw.collectInstrumentTypes()

	return meta
}

// countComponentsAndSamples applies the variant-specific counting rules.
func (raw *rawDrumkit) countComponentsAndSamples() (components int, samples int) {
	switch {
	case len(raw.componentList) > 0:
		// v1.2.3: components declared at top level; count from the list.
		components = len(raw.componentList)
		for _, instr := range raw.instruments {
			for _, ic := range instr.instrComponents {
				samples += ic.layers
			}
		}

	case raw.hasInstrumentComponents():
		// v2.0.0: components inferred from unique names across all instruments.
		seen := make(map[string]struct{})
		for _, instr := range raw.instruments {
			for _, ic := range instr.instrComponents {
				if _, exists := seen[ic.name]; !exists {
					seen[ic.name] = struct{}{}
					components++
				}
				samples += ic.layers
			}
		}

	default:
		// Legacy formats: layers directly under instrument, no components.
		components = 0
		for _, instr := range raw.instruments {
			samples += instr.directLayers
		}
	}
	return
}

// hasInstrumentComponents reports whether any instrument uses <instrumentComponent> wrappers.
func (raw *rawDrumkit) hasInstrumentComponents() bool {
	for _, instr := range raw.instruments {
		if len(instr.instrComponents) > 0 {
			return true
		}
	}
	return false
}

// collectInstrumentTypes returns a deduplicated list of non-empty type strings.
func (raw *rawDrumkit) collectInstrumentTypes() []string {
	seen := make(map[string]struct{})
	var types []string
	for _, instr := range raw.instruments {
		t := strings.TrimSpace(instr.instrumentType)
		if t == "" {
			continue
		}
		if _, exists := seen[t]; !exists {
			seen[t] = struct{}{}
			types = append(types, t)
		}
	}
	return nilToEmpty(types)
}

// seekElement advances the decoder until it finds a start element with the
// given local name, returning an error if EOF is reached first.
func seekElement(dec *xml.Decoder, name string) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if se, ok := tok.(xml.StartElement); ok && localName(se) == name {
			return nil
		}
	}
}

// skipElement consumes all tokens up to and including the end element that
// closes the element whose start was already consumed.
func skipElement(dec *xml.Decoder, name string) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if localName(t) == name {
				depth++
			}
		case xml.EndElement:
			if localName(t) == name {
				depth--
			}
		}
	}
	return nil
}

// charData reads the text content of the current element and consumes its
// closing tag.  It accumulates all CharData tokens, ignoring nested elements.
func charData(dec *xml.Decoder, name string) (string, error) {
	var buf strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			// Nested elements inside text content (rare but possible) — skip them.
			if err := skipElement(dec, localName(t)); err != nil {
				return "", err
			}
		case xml.EndElement:
			_ = name // end element matches what was opened by our caller
			return buf.String(), nil
		}
	}
}

// intCharData reads an integer from the text content of the current element.
func intCharData(dec *xml.Decoder, name string) (int, error) {
	s, err := charData(dec, name)
	if err != nil {
		return 0, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return 0, fmt.Errorf("expected integer, got %q: %w", s, err)
	}
	return v, nil
}

// localName returns the local part of an XML element's name, stripping any namespace.
func localName[T xml.StartElement | xml.EndElement](t T) string {
	switch v := any(t).(type) {
	case xml.StartElement:
		return v.Name.Local
	case xml.EndElement:
		return v.Name.Local
	}
	return ""
}

// nilToEmpty ensures callers always receive a non-nil slice for JSON marshalling.
func nilToEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
