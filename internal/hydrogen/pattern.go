package hydrogen

import "github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"

func ParsePatternFile(path string, options ParseOptions) (indexfile.PatternArtifact, error) {
	root, err := parseXMLFile(path)
	if err != nil {
		return indexfile.PatternArtifact{}, err
	}

	pattern := root.child("pattern")
	if pattern == nil {
		return indexfile.PatternArtifact{}, errUnexpectedRoot(path, root.XMLName.Local, "pattern")
	}

	artifact := indexfile.PatternArtifact{Artifact: baseArtifact(indexfile.ArtifactTypePattern, options)}
	artifact.Name = firstNonEmpty(pattern.text("name"), pattern.text("pattern_name"))
	artifact.Author = firstNonEmpty(root.text("author"), pattern.text("author"))
	artifact.Description = pattern.text("info")
	artifact.Version = parseIntOrDefault(pattern.text("userVersion"), 0)
	artifact.FormatVersion = parseIntOrDefault(pattern.text("formatVersion"), 1)
	artifact.License = firstNonEmpty(root.text("license"), pattern.text("license"))
	artifact.Tags = tagsFromPatternRoot(*pattern)
	artifact.Notes = len(nodesAtPath(*pattern, "noteList", "note"))
	artifact.InstrumentTypes = uniqueSorted(collectTexts(*pattern, "noteList", "note", "type"))

	return artifact, nil
}

func tagsFromPatternRoot(pattern xmlNode) []string {
	tags := uniqueSorted(collectTexts(pattern, "tags", "tag"))
	if len(tags) > 0 {
		return tags
	}

	category := pattern.text("category")
	if category == "" || category == "unknown" || category == "not_categorized" {
		return []string{}
	}

	return []string{category}
}
