package compiler

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolveSpecPath(path string) (string, error) {
	candidates := []string{
		path,
		filepath.Join("..", "..", path),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("dagruntime workflow spec not found: %s", path)
}
