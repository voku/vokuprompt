package cli

import (
	"os"
	"path/filepath"
)

var executablePath = os.Executable

func resolveAssetPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if _, err := os.Stat(path); err == nil {
		return path
	}

	executable, err := executablePath()
	if err != nil {
		return path
	}

	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}

	candidate := filepath.Join(filepath.Dir(executable), path)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	return path
}
