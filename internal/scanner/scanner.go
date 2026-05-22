// Package scanner walks a directory tree and discovers Hydrogen artifact files,
// parsing their metadata and computing file hashes.
package scanner

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hydrogen-music/hydrogen-index/internal/parser"
)

// ArtifactFile represents a discovered artifact with its parsed metadata and file info.
type ArtifactFile struct {
	Path     string      // absolute path to the file
	RelPath  string      // relative path from scan root
	BaseURL  string      // base URL prepended to RelPath for full permalink
	Hash     string      // hex-encoded SHA-256
	Size     int64
	Metadata interface{} // one of *model.DrumkitMetadata, *model.PatternMetadata, *model.SongMetadata
}

// Scan walks the directory tree rooted at dir and returns all discovered Hydrogen artifacts.
// Non-fatal parse errors are collected rather than stopping the scan.
// The baseURL parameter is prepended to each artifact's relative path when
// constructing the full URL in the index.
func Scan(dir, baseURL string) ([]ArtifactFile, []error) {
	var results []ArtifactFile
	var errs []error

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Collect directory-access errors rather than aborting the entire walk.
			errs = append(errs, fmt.Errorf("walk %s: %w", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}

		artifact, parseErr := processFile(path, dir)
		if parseErr != nil {
			errs = append(errs, parseErr)
			return nil
		}
		if artifact != nil {
			artifact.BaseURL = baseURL
			results = append(results, *artifact)
		}
		return nil
	})
	if walkErr != nil {
		errs = append(errs, fmt.Errorf("walk root: %w", walkErr))
	}

	return results, errs
}

// processFile inspects a single file path and returns an ArtifactFile if it is
// a recognised Hydrogen artifact, nil if it should be skipped, or an error if
// it looks like an artifact but could not be parsed.
func processFile(path, root string) (*ArtifactFile, error) {
	ext := filepath.Ext(path)

	switch {
	case ext == ".h2pattern":
		return parseRegular(path, root, func(r io.Reader) (interface{}, error) {
			return parser.ParsePattern(r)
		})

	case ext == ".h2song":
		return parseRegular(path, root, func(r io.Reader) (interface{}, error) {
			return parser.ParseSong(r)
		})

	case ext == ".h2drumkit":
		return parseDrumkitTar(path, root)
	}

	return nil, nil
}

// parseRegular reads the file at path into memory, computes its hash and size,
// then passes the bytes to parse for metadata extraction.
func parseRegular(path, root string, parse func(io.Reader) (interface{}, error)) (*ArtifactFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	meta, err := parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	relPath, err := relURL(root, path)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(data)
	return &ArtifactFile{
		Path:     path,
		RelPath:  relPath,
		Hash:     hex.EncodeToString(sum[:]),
		Size:     int64(len(data)),
		Metadata: meta,
	}, nil
}

// parseDrumkitTar handles .h2drumkit files, which are tar archives containing
// a drumkit.xml at an arbitrary depth. The hash and size are of the archive
// itself, not the extracted XML entry.
func parseDrumkitTar(path, root string) (*ArtifactFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	meta, err := findAndParseDrumkitXML(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse drumkit tar %s: %w", path, err)
	}

	relPath, err := relURL(root, path)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(data)
	return &ArtifactFile{
		Path:     path,
		RelPath:  relPath,
		Hash:     hex.EncodeToString(sum[:]),
		Size:     int64(len(data)),
		Metadata: meta,
	}, nil
}

// findAndParseDrumkitXML walks tar entries, validates that all entries reside
// under exactly one top-level folder, extracts drumkit.xml, and parses it.
// The folder name is stored in the returned DrumkitMetadata.FolderName.
//
// Validation rules:
//   - Every entry must be inside a top-level folder (no root-level files)
//   - There must be exactly one top-level folder across all entries
//   - drumkit.xml must be present somewhere in the archive
func findAndParseDrumkitXML(r io.Reader) (interface{}, error) {
	tr := tar.NewReader(r)
	topLevelFolders := make(map[string]struct{})
	var drumkitData []byte

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		// Determine the top-level folder for this entry.
		slashIdx := strings.Index(hdr.Name, "/")
		if slashIdx == -1 {
			// Entry is at root level (no containing folder).
			return nil, fmt.Errorf("archive contains top-level entry %q; expected all entries within a single top-level folder", hdr.Name)
		}
		topLevelFolders[hdr.Name[:slashIdx]] = struct{}{}

		// Buffer drumkit.xml content for later parsing.
		if filepath.Base(hdr.Name) == "drumkit.xml" && hdr.Typeflag == tar.TypeReg {
			drumkitData, err = io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("reading drumkit.xml: %w", err)
			}
		}
	}

	// Validate exactly one top-level folder.
	if len(topLevelFolders) == 0 {
		return nil, fmt.Errorf("archive is empty")
	}
	if len(topLevelFolders) > 1 {
		folders := make([]string, 0, len(topLevelFolders))
		for name := range topLevelFolders {
			folders = append(folders, name)
		}
		return nil, fmt.Errorf("archive contains %d top-level folders (%s); expected exactly one",
			len(topLevelFolders), strings.Join(folders, ", "))
	}

	// Extract the folder name.
	var folderName string
	for name := range topLevelFolders {
		folderName = name
	}

	if drumkitData == nil {
		return nil, fmt.Errorf("drumkit.xml not found in archive")
	}

	meta, err := parser.ParseDrumkit(bytes.NewReader(drumkitData))
	if err != nil {
		return nil, err
	}
	meta.FolderName = folderName
	return meta, nil
}

// relURL computes the URL-style relative path from root to path.
// filepath.ToSlash ensures forward slashes regardless of platform.
func relURL(root, path string) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", fmt.Errorf("rel path from %s to %s: %w", root, path, err)
	}
	return filepath.ToSlash(rel), nil
}
