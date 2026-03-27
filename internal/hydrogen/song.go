package hydrogen

import "github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"

func ParseSongFile(path string, options ParseOptions) (indexfile.SongArtifact, error) {
	root, err := parseXMLFile(path)
	if err != nil {
		return indexfile.SongArtifact{}, err
	}

	artifact := indexfile.SongArtifact{Artifact: baseArtifact(indexfile.ArtifactTypeSong, options)}
	artifact.Name = root.text("name")
	artifact.Author = root.text("author")
	artifact.Description = root.text("notes")
	artifact.Version = parseIntOrDefault(root.text("userVersion"), 0)
	artifact.FormatVersion = parseIntOrDefault(root.text("formatVersion"), 1)
	artifact.Tags = uniqueSorted(collectTexts(root, "tags", "tag"))
	artifact.License = root.text("license")
	artifact.Patterns = len(nodesAtPath(root, "patternList", "pattern"))

	return artifact, nil
}
