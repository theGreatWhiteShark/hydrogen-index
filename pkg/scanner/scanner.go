package scanner

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FindGitRoot traverses upwards from currentDir until a directory containing a .git folder is found.
func FindGitRoot(currentDir string) (string, error) {
	abs, err := filepath.Abs(currentDir)
	if err != nil {
		return "", err
	}

	curr := abs
	for {
		gitPath := filepath.Join(curr, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", fmt.Errorf("git root not found starting from %s", currentDir)
}

// ScanArtifacts recursively traverses dir and returns a list of files ending with .h2pattern, .h2song, or .h2drumkit.
func ScanArtifacts(dir string) ([]string, error) {
	var artifacts []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".h2pattern", ".h2song", ".h2drumkit":
			artifacts = append(artifacts, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

// ExtractDrumkitXML extracts drumkit.xml from a .h2drumkit tar archive to a temporary file.
// The caller is responsible for removing the temporary file.
func ExtractDrumkitXML(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if filepath.Base(header.Name) == "drumkit.xml" {
			tmpFile, err := os.CreateTemp("", "drumkit-*.xml")
			if err != nil {
				return "", err
			}
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, tr); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}
			return tmpFile.Name(), nil
		}
	}

	return "", errors.New("drumkit.xml not found in archive")
}
