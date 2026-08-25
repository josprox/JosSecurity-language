// Package diagnostics defines the common diagnostic model used by the Joss
// parser, semantic analyzer and command-line tooling.
package diagnostics

import "fmt"

// Severity describes how a diagnostic affects program validity.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Position is a one-based source position. A zero column means that the
// producing phase only knows the line.
type Position struct {
	Line   int
	Column int
}

// Range identifies a source span. End may equal Start when only one position
// is known.
type Range struct {
	Start Position
	End   Position
}

// Diagnostic is the single interchange format for all non-runtime findings.
type Diagnostic struct {
	Code        string
	Severity    Severity
	Message     string
	File        string
	Range       Range
	Explanation string
	Suggestion  string
}

// String returns the stable human-readable representation used by the CLI.
func (d Diagnostic) String() string {
	location := d.File
	if d.Range.Start.Line > 0 {
		if location != "" {
			location += ":"
		}
		location += fmt.Sprintf("%d", d.Range.Start.Line)
		if d.Range.Start.Column > 0 {
			location += fmt.Sprintf(":%d", d.Range.Start.Column)
		}
	}
	if location != "" {
		location += ": "
	}
	return fmt.Sprintf("%s[%s] %s%s", d.Severity, d.Code, location, d.Message)
}

// Bag accumulates diagnostics while preserving their structured form.
type Bag struct {
	items []Diagnostic
}

func (b *Bag) Add(d Diagnostic) { b.items = append(b.items, d) }

func (b *Bag) Items() []Diagnostic {
	result := make([]Diagnostic, len(b.items))
	copy(result, b.items)
	return result
}

func (b *Bag) HasErrors() bool {
	for _, item := range b.items {
		if item.Severity == SeverityError {
			return true
		}
	}
	return false
}
