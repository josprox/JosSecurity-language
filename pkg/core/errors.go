package core

import (
	"fmt"
	"strings"
)

// JossError represents a structured runtime error in Joss.
// It carries source location information for precise error reporting.
type JossError struct {
	Type    string // "UndefinedVariable", "UndefinedFunction", "UndefinedClass", etc.
	Message string
	File    string
	Line    int
	Column  int
	Source  string // Source line content for snippet display
}

// Error implements the error interface with a detailed, structured output.
func (e *JossError) Error() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Error: %s\n", e.Message))

	if e.File != "" {
		b.WriteString(fmt.Sprintf("\nFile: %s\n", e.File))
	}
	if e.Line > 0 {
		b.WriteString(fmt.Sprintf("Line: %d\n", e.Line))
	}
	if e.Column > 0 {
		b.WriteString(fmt.Sprintf("Column: %d\n", e.Column))
	}

	if e.Source != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", e.Source))
		if e.Column > 0 {
			b.WriteString(strings.Repeat(" ", e.Column-1))
			b.WriteString("^\n")
		}
	}

	return b.String()
}

// NewJossError creates a JossError with the given type and message.
func NewJossError(errType, message, file string, line int) *JossError {
	return &JossError{
		Type:    errType,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// IsJossError checks whether a recovered panic value is a JossError.
func IsJossError(v interface{}) bool {
	_, ok := v.(*JossError)
	return ok
}

// FormatPanicAsError converts any panic value to a presentable error string.
// It preserves JossError formatting and wraps other panic types.
func FormatPanicAsError(v interface{}) string {
	if je, ok := v.(*JossError); ok {
		return je.Error()
	}
	if err, ok := v.(error); ok {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}
	return fmt.Sprintf("Error: %v\n", v)
}
