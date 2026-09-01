// Package errors defines structured runtime failures without depending on the
// AST evaluator, stdlib or web framework.
package errors

import (
	"fmt"
	"strings"
)

type Frame struct {
	Function string
	Class    string
	File     string
	Line     int
	Column   int
}

// JossError is the common runtime error envelope. Type remains for source/API
// compatibility; Code is the stable machine-readable vocabulary.
type JossError struct {
	Code       string
	Type       string
	Message    string
	File       string
	Line       int
	Column     int
	Source     string
	StackTrace []Frame
	Context    map[string]interface{}
	Suggestion string
	Cause      error
}

func (e *JossError) Error() string {
	if e == nil {
		return "Error: <nil>\n"
	}
	var b strings.Builder
	b.WriteString("Error")
	if e.Code != "" {
		b.WriteString("[")
		b.WriteString(e.Code)
		b.WriteString("]")
	}
	b.WriteString(": ")
	b.WriteString(e.Message)
	b.WriteString("\n")
	if e.File != "" {
		fmt.Fprintf(&b, "\nFile: %s\n", e.File)
	}
	if e.Line > 0 {
		fmt.Fprintf(&b, "Line: %d\n", e.Line)
	}
	if e.Column > 0 {
		fmt.Fprintf(&b, "Column: %d\n", e.Column)
	}
	if e.Source != "" {
		fmt.Fprintf(&b, "\n%s\n", e.Source)
		if e.Column > 0 {
			b.WriteString(strings.Repeat(" ", e.Column-1))
			b.WriteString("^\n")
		}
	}
	if len(e.StackTrace) > 0 {
		b.WriteString("\nJoss stack:\n")
		for index := len(e.StackTrace) - 1; index >= 0; index-- {
			frame := e.StackTrace[index]
			name := frame.Function
			if frame.Class != "" {
				name = frame.Class + "::" + name
			}
			fmt.Fprintf(&b, "  at %s", name)
			if frame.File != "" {
				fmt.Fprintf(&b, " (%s", frame.File)
				if frame.Line > 0 {
					fmt.Fprintf(&b, ":%d", frame.Line)
				}
				b.WriteString(")")
			}
			b.WriteByte('\n')
		}
	}
	if e.Suggestion != "" {
		fmt.Fprintf(&b, "\nSuggestion: %s\n", e.Suggestion)
	}
	if e.Cause != nil {
		fmt.Fprintf(&b, "\nCaused by: %v\n", e.Cause)
	}
	return b.String()
}

func (e *JossError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// AttachStack records the first (deepest) complete Joss stack and leaves it
// unchanged while the same error bubbles through outer frames.
func (e *JossError) AttachStack(frames []Frame) {
	if e == nil || len(e.StackTrace) != 0 || len(frames) == 0 {
		return
	}
	e.StackTrace = append([]Frame(nil), frames...)
}
