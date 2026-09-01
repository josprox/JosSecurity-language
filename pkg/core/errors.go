package core

import (
	"fmt"

	runtimeerrors "github.com/jossecurity/joss/pkg/runtime/errors"
)

// JossError remains available from core for compatibility while its
// implementation lives in the language runtime layer.
type JossError = runtimeerrors.JossError
type JossStackFrame = runtimeerrors.Frame

// NewJossError creates a JossError with the given type and message.
func NewJossError(errType, message, file string, line int) *JossError {
	return &JossError{
		Type:    errType,
		Message: message,
		File:    file,
		Line:    line,
	}
}

func IsJossError(value interface{}) bool {
	_, ok := value.(*JossError)
	return ok
}

// FormatPanicAsError converts any panic value to a presentable error string.
// It preserves JossError formatting and wraps other panic types.
func FormatPanicAsError(value interface{}) string {
	if err, ok := value.(*JossError); ok {
		return err.Error()
	}
	if err, ok := value.(error); ok {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}
	return fmt.Sprintf("Error: %v\n", value)
}
