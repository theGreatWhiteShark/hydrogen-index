package hydrogen

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
)

func ParseDrumkitArchive(path string, options ParseOptions) (indexfile.DrumkitArtifact, error) {
	temporaryDir, err := os.MkdirTemp("", "hydrogen-index-drumkit-")
	if err != nil {
		return indexfile.DrumkitArtifact{}, fmt.Errorf("create temporary extraction directory: %w", err)
	}
	defer os.RemoveAll(temporaryDir)

	// Extracting to a temporary directory keeps archive traversal logic isolated
	// from XML parsing and mirrors the archive layout users distribute.
	if err := extractTarArchive(path, temporaryDir); err != nil {
		return indexfile.DrumkitArtifact{}, err
	}

	drumkitPath, err := findDrumkitXML(temporaryDir)
	if err != nil {
		return indexfile.DrumkitArtifact{}, err
	}

	return ParseDrumkitXMLFile(drumkitPath, options)
}

func ParseDrumkitXMLFile(path string, options ParseOptions) (indexfile.DrumkitArtifact, error) {
	root, err := parseXMLFile(path)
	if err != nil {
		return indexfile.DrumkitArtifact{}, err
	}

	if root.XMLName.Local != "drumkit_info" {
		return indexfile.DrumkitArtifact{}, errUnexpectedRoot(path, root.XMLName.Local, "drumkit_info")
	}

	artifact := indexfile.DrumkitArtifact{Artifact: baseArtifact(indexfile.ArtifactTypeDrumkit, options)}
	artifact.Name = root.text("name")
	artifact.Author = root.text("author")
	artifact.Description = root.text("info")
	artifact.Version = parseIntOrDefault(root.text("userVersion"), 0)
	artifact.FormatVersion = parseIntOrDefault(root.text("formatVersion"), 1)
	artifact.Tags = uniqueSorted(collectTexts(root, "tags", "tag"))
	artifact.License = root.text("license")
	artifact.Instruments = len(nodesAtPath(root, "instrumentList", "instrument"))

	drumkitComponents := len(nodesAtPath(root, "componentList", "drumkitComponent"))
	instrumentComponents := len(nodesAtPath(root, "instrumentList", "instrument", "instrumentComponent"))
	if drumkitComponents > 0 {
		artifact.Components = drumkitComponents
	} else {
		artifact.Components = instrumentComponents
	}

	artifact.Samples = len(collectLayerNodes(root))
	artifact.InstrumentTypes = uniqueSorted(instrumentTypes(root))

	return artifact, nil
}

func collectLayerNodes(root xmlNode) []xmlNode {
	var layers []xmlNode
	for _, instrument := range nodesAtPath(root, "instrumentList", "instrument") {
		layers = append(layers, nodesAtPath(instrument, "instrumentComponent", "layer")...)
		layers = append(layers, instrument.children("layer")...)
	}
	return layers
}

func instrumentTypes(root xmlNode) []string {
	var values []string
	for _, instrument := range nodesAtPath(root, "instrumentList", "instrument") {
		instrumentType := instrument.text("type")
		if instrumentType != "" {
			values = append(values, instrumentType)
			continue
		}
		values = append(values, instrument.text("name"))
	}
	return values
}

func extractTarArchive(path string, destinationDir string) error {
	archiveFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer archiveFile.Close()

	reader := tar.NewReader(archiveFile)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		targetPath, err := safeJoin(destinationDir, header.Name)
		if err != nil {
			return fmt.Errorf("extract %q: %w", path, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %q: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create parent directory for %q: %w", targetPath, err)
			}

			outputFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create %q: %w", targetPath, err)
			}

			if _, err := io.Copy(outputFile, reader); err != nil {
				_ = outputFile.Close()
				return fmt.Errorf("write %q: %w", targetPath, err)
			}

			if err := outputFile.Close(); err != nil {
				return fmt.Errorf("close %q: %w", targetPath, err)
			}
		}
	}
}

func safeJoin(root string, archivePath string) (string, error) {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(cleanRoot, filepath.FromSlash(archivePath))
	cleanTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	if cleanTarget != cleanRoot && !strings.HasPrefix(cleanTarget, cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry %q escapes temporary directory", archivePath)
	}

	return cleanTarget, nil
}

func findDrumkitXML(root string) (string, error) {
	var drumkitPath string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(entry.Name(), "drumkit.xml") {
			drumkitPath = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("search extracted drumkit metadata: %w", err)
	}
	if drumkitPath == "" {
		return "", fmt.Errorf("archive %q does not contain drumkit.xml", root)
	}
	return drumkitPath, nil
}
