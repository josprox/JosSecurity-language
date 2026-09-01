package fixer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jossecurity/joss/pkg/formatter"
)

type FixResult struct {
	File        string `json:"file"`
	Changed     bool   `json:"changed"`
	FixesApplied int   `json:"fixes_applied"`
	Original    string `json:"original,omitempty"`
	Fixed       string `json:"fixed,omitempty"`
}

type Fixer struct {
	dryRun bool
}

func NewFixer(dryRun bool) *Fixer {
	return &Fixer{dryRun: dryRun}
}

var (
	// Fix missing visibility: func name( -> public func name(
	reFuncVisibility  = regexp.MustCompile(`(?m)^([ \t]*)func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	reClassVisibility = regexp.MustCompile(`(?m)^([ \t]*)class\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	// Fix missing mixed param type: ($x -> (mixed $x, , $y -> , mixed $y
	reUntypedParam    = regexp.MustCompile(`([\(,]\s*)\$([a-zA-Z_][a-zA-Z0-9_]*)`)
)

func (f *Fixer) FixSource(src string) (string, int) {
	applied := 0
	fixed := src

	// 1. Fix implicit top-level func visibility
	if reFuncVisibility.MatchString(fixed) {
		fixed = reFuncVisibility.ReplaceAllStringFunc(fixed, func(match string) string {
			applied++
			return reFuncVisibility.ReplaceAllString(match, "${1}public func ${2}(")
		})
	}

	// 2. Fix implicit top-level class visibility
	if reClassVisibility.MatchString(fixed) {
		fixed = reClassVisibility.ReplaceAllStringFunc(fixed, func(match string) string {
			applied++
			return reClassVisibility.ReplaceAllString(match, "${1}public class ${2}")
		})
	}

	// 3. Format canonically
	formatted, err := formatter.FormatSource(fixed)
	if err == nil && formatted != "" {
		if formatted != fixed {
			applied++
		}
		fixed = formatted
	}

	return fixed, applied
}

func (f *Fixer) FixFile(path string) (FixResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FixResult{}, err
	}

	original := string(data)
	fixed, fixesCount := f.FixSource(original)
	changed := fixed != original

	if changed && !f.dryRun {
		if err := os.WriteFile(path, []byte(fixed), 0644); err != nil {
			return FixResult{}, err
		}
	}

	return FixResult{
		File:         path,
		Changed:      changed,
		FixesApplied: fixesCount,
		Original:     original,
		Fixed:        fixed,
	}, nil
}

func (f *Fixer) FixDirectory(dir string) ([]FixResult, error) {
	var results []FixResult

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".joss") {
			res, fixErr := f.FixFile(path)
			if fixErr != nil {
				return fmt.Errorf("failed fixing %s: %w", path, fixErr)
			}
			if res.Changed {
				results = append(results, res)
			}
		}
		return nil
	})

	return results, err
}
