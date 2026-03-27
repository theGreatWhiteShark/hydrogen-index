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

	"github.com/hydrogen-music/hydrogen-index/internal/parser"
)

// ArtifactFile represents a discovered artifact with its parsed metadata and file info.
type ArtifactFile struct {
	Path     string      // absolute path to the file
	RelPath  string      // relative path from scan root (used as URL in index)
	Hash     string      // hex-encoded SHA-256
	Size     int64
	Metadata interface{} // one of *model.DrumkitMetadata, *model.PatternMetadata, *model.SongMetadata
}

// Scan walks the directory tree rooted at dir and returns all discovered Hydrogen artifacts.
// Non-fatal parse errors are collected rather than stopping the scan.
func Scan(dir string) ([]ArtifactFile, []error) {
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
	base := filepath.Base(path)
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

	case base == "drumkit.xml":
		return parseRegular(path, root, func(r io.Reader) (interface{}, error) {
			return parser.ParseDrumkit(r)
		})
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

// findAndParseDrumkitXML walks tar entries looking for drumkit.xml at any depth
// and parses it. The first matching entry wins.
func findAndParseDrumkitXML(r io.Reader) (interface{}, error) {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("drumkit.xml not found in archive")
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}
		if filepath.Base(hdr.Name) == "drumkit.xml" {
			return parser.ParseDrumkit(tr)
		}
	}
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
