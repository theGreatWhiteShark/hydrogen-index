package scan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/gitrepo"
	"github.com/theGreatWhiteShark/hydrogen-index/internal/hydrogen"
	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
)

type Options struct {
	WorkingDir string
	Directory  string
	OutputPath string
	BaseURL    string
	Version    string
	Now        func() time.Time
}

func Run(options Options) error {
	if options.Now == nil {
		options.Now = time.Now
	}

	scanRoot, err := resolveScanRoot(options)
	if err != nil {
		return err
	}

	outputPath, err := resolveOutputPath(options.WorkingDir, options.OutputPath)
	if err != nil {
		return err
	}

	document, err := buildDocument(scanRoot, options.BaseURL, options.Version, options.Now())
	if err != nil {
		return err
	}

	data, err := indexfile.Marshal(document)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory for %q: %w", outputPath, err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", outputPath, err)
	}

	return nil
}

func resolveScanRoot(options Options) (string, error) {
	if options.Directory != "" {
		return absoluteFromWorkingDir(options.WorkingDir, options.Directory)
	}

	return gitrepo.FindRoot(options.WorkingDir)
}

func resolveOutputPath(workingDir string, outputPath string) (string, error) {
	if outputPath == "" {
		return filepath.Join(workingDir, "index.json"), nil
	}

	return absoluteFromWorkingDir(workingDir, outputPath)
}

func absoluteFromWorkingDir(workingDir string, target string) (string, error) {
	if filepath.IsAbs(target) {
		return filepath.Clean(target), nil
	}

	absolutePath := filepath.Join(workingDir, target)
	return filepath.Abs(absolutePath)
}

func buildDocument(scanRoot string, baseURL string, version string, createdAt time.Time) (indexfile.Document, error) {
	var patterns []indexfile.PatternArtifact
	var songs []indexfile.SongArtifact
	var drumkits []indexfile.DrumkitArtifact

	err := filepath.WalkDir(scanRoot, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, err := filepath.Rel(scanRoot, filePath)
		if err != nil {
			return fmt.Errorf("derive path for %q: %w", filePath, err)
		}

		artifactURL, err := deriveArtifactURL(baseURL, relativePath)
		if err != nil {
			return err
		}

		parseOptions, err := buildParseOptions(filePath, artifactURL)
		if err != nil {
			return err
		}

		switch {
		case isPatternFile(entry.Name()):
			artifact, err := hydrogen.ParsePatternFile(filePath, parseOptions)
			if err != nil {
				return err
			}
			patterns = append(patterns, artifact)
		case isSongFile(entry.Name()):
			artifact, err := hydrogen.ParseSongFile(filePath, parseOptions)
			if err != nil {
				return err
			}
			songs = append(songs, artifact)
		case isDrumkitArchive(entry.Name()):
			artifact, err := hydrogen.ParseDrumkitArchive(filePath, parseOptions)
			if err != nil {
				return err
			}
			drumkits = append(drumkits, artifact)
		case isLegacyDrumkitXML(entry.Name()):
			artifact, err := hydrogen.ParseDrumkitXMLFile(filePath, parseOptions)
			if err != nil {
				return err
			}
			drumkits = append(drumkits, artifact)
		}

		return nil
	})
	if err != nil {
		return indexfile.Document{}, fmt.Errorf("scan %q: %w", scanRoot, err)
	}

	sortArtifacts(patterns, songs, drumkits)

	return indexfile.Document{
		Version:      version,
		Created:      createdAt.Format("2006-01-02T15:04:05"),
		PatternCount: len(patterns),
		SongCount:    len(songs),
		DrumkitCount: len(drumkits),
		Patterns:     patterns,
		Songs:        songs,
		Drumkits:     drumkits,
	}, nil
}

func buildParseOptions(path string, artifactURL string) (hydrogen.ParseOptions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return hydrogen.ParseOptions{}, fmt.Errorf("read %q: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return hydrogen.ParseOptions{}, fmt.Errorf("stat %q: %w", path, err)
	}

	sum := sha256.Sum256(data)
	return hydrogen.ParseOptions{
		URL:  artifactURL,
		Size: info.Size(),
		Hash: hex.EncodeToString(sum[:]),
	}, nil
}

func deriveArtifactURL(baseURL string, relativePath string) (string, error) {
	normalizedRelativePath := filepath.ToSlash(relativePath)
	if baseURL == "" {
		return normalizedRelativePath, nil
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL %q: %w", baseURL, err)
	}

	segments := strings.Split(normalizedRelativePath, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}

	joinedPath := strings.TrimSuffix(parsedURL.Path, "/") + "/" + path.Join(segments...)
	parsedURL.Path = joinedPath
	parsedURL.RawPath = joinedPath

	return parsedURL.String(), nil
}

func isPatternFile(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".h2pattern")
}

func isSongFile(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".h2song")
}

func isDrumkitArchive(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".h2drumkit")
}

func isLegacyDrumkitXML(name string) bool {
	return strings.EqualFold(name, "drumkit.xml")
}

func sortArtifacts(patterns []indexfile.PatternArtifact, songs []indexfile.SongArtifact, drumkits []indexfile.DrumkitArtifact) {
	slices.SortFunc(patterns, func(left indexfile.PatternArtifact, right indexfile.PatternArtifact) int {
		return strings.Compare(left.URL, right.URL)
	})
	slices.SortFunc(songs, func(left indexfile.SongArtifact, right indexfile.SongArtifact) int {
		return strings.Compare(left.URL, right.URL)
	})
	slices.SortFunc(drumkits, func(left indexfile.DrumkitArtifact, right indexfile.DrumkitArtifact) int {
		return strings.Compare(left.URL, right.URL)
	})
}
