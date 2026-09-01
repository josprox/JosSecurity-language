package parser

import (
	"path/filepath"
	"strings"
)

// EnvironmentFileNames contains standard environment file names.
var EnvironmentFileNames = []string{
	"env.joss",
	".env",
	"joss.env",
	"env.enc",
	".env.enc",
	"env.json",
	".env.json",
}

// IsEnvFile returns true if the file path points to an environment, config or encrypted environment file.
func IsEnvFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".enc" {
		return true
	}

	for _, name := range EnvironmentFileNames {
		if base == name {
			return true
		}
	}

	if strings.HasPrefix(base, ".env.") || strings.HasPrefix(base, "env.") {
		return true
	}

	return false
}

// IsIgnoredSourceFile returns true if the given file should not be processed as executable Joss source code.
func IsIgnoredSourceFile(path string) bool {
	return IsEnvFile(path)
}
