package gitrepo

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path for %q: %w", startPath, err)
	}

	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect %q: %w", gitPath, err)
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("%q is not inside a git repository", startPath)
		}

		currentPath = parentPath
	}
}
