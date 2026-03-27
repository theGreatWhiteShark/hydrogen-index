package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/hydrogen-music/hydrogen-index/pkg/domain"
	"github.com/hydrogen-music/hydrogen-index/pkg/parser"
	"github.com/hydrogen-music/hydrogen-index/pkg/scanner"
)

// BuildIndex scans a directory for Hydrogen artifacts and builds an index file.
func BuildIndex(scanDir string, outputFilePath string) error {
	artifacts, err := scanner.ScanArtifacts(scanDir)
	if err != nil {
		return err
	}

	index := domain.IndexFile{
		Version: "1.0.0",
		Created: time.Now().Format(time.RFC3339),
	}

	for _, artifactPath := range artifacts {
		ext := filepath.Ext(artifactPath)
		var xmlPath string
		var isTemp bool

		if ext == ".h2drumkit" {
			xmlPath, err = scanner.ExtractDrumkitXML(artifactPath)
			if err != nil {
				// Skip or handle error
				continue
			}
			isTemp = true
		} else {
			xmlPath = artifactPath
		}

		parsed, err := parser.ParseArtifact(xmlPath, artifactPath)
		if isTemp {
			os.Remove(xmlPath)
		}

		if err != nil {
			continue
		}

		switch v := parsed.(type) {
		case *domain.PatternBlock:
			index.Patterns = append(index.Patterns, *v)
		case *domain.SongBlock:
			index.Songs = append(index.Songs, *v)
		case *domain.DrumkitBlock:
			index.Drumkits = append(index.Drumkits, *v)
		}
	}

	index.PatternCount = len(index.Patterns)
	index.SongCount = len(index.Songs)
	index.DrumkitCount = len(index.Drumkits)

	// Sort for determinism
	sort.Slice(index.Patterns, func(i, j int) bool { return index.Patterns[i].Name < index.Patterns[j].Name })
	sort.Slice(index.Songs, func(i, j int) bool { return index.Songs[i].Name < index.Songs[j].Name })
	sort.Slice(index.Drumkits, func(i, j int) bool { return index.Drumkits[i].Name < index.Drumkits[j].Name })

	// Calculate hash
	index.Hash = ""
	jsonBytes, err := json.Marshal(index)
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write(jsonBytes)
	index.Hash = hex.EncodeToString(h.Sum(nil))

	// Re-marshal with hash
	finalBytes, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputFilePath, finalBytes, 0644)
}
