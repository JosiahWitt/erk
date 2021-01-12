package erk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/JosiahWitt/erk/erkstrict"
)

// Error satisfies the Erkable interface.
var (
	_ Erkable         = &Error{}
	_ ErrorIndentable = &Error{}
)

// Error stores details about an error with kinds and a message template.
type Error struct {
	kind    Kind
	message string
	params  Params
}

// ExportedError that can be used outside the erk package.
// A common use case is marshalling the error to JSON.
type ExportedError struct {
	BaseExport
}

// New creates an error with a kind and message.
func New(kind Kind, message string) error {
	return NewWith(kind, message, nil)
}

// NewWith creates an error with a kind, message, and params.
func NewWith(kind Kind, message string, params Params) error {
	e := &Error{
		kind:    kind,
		message: message,
		params:  params,
	}

	// If strict mode, ensure we can parse the template
	if erkstrict.IsStrictMode() {
		e.parseTemplate() //nolint:errcheck // Panics if there is an error
	}

	return e
}

// Error processes the message template with the provided params.
func (e *Error) Error() string {
	return e.IndentError(IndentSpaces)
}

// IndentError processes the message template with the provided params and indentation.
//
// The indentLevel represents the indentation of wrapped errors.
// Thus, it should start with "  ".
func (e *Error) IndentError(indentLevel string) string {
	t, err := e.parseTemplate()
	if err != nil {
		return e.message
	}

	if erkstrict.IsStrictMode() {
		t.Option("missingkey=error")
	}

	var filledMessage bytes.Buffer
	err = t.Execute(&filledMessage, e.params.prep(indentLevel))
	if err != nil {
		if erkstrict.IsStrictMode() {
			panic(fmt.Sprintf("Unable to execute error template:\n\tKind: %s\n\tTemplate: %s\n\tParams: %+v\n\tError: %v\n",
				GetKindString(e),
				e.message,
				e.params,
				err,
			))
		}

		return e.message
	}

	return filledMessage.String()
}

// Is implements the Go 1.13+ Is interface for use with errors.Is.
func (e *Error) Is(err error) bool {
	// Allows validating the error when comparing errors during testing
	if erkstrict.IsStrictMode() {
		e.Error() //nolint:govet // Panics if there is an error
	}

	var e2 *Error
	if errors.As(err, &e2) {
		return IsKind(err, e.kind) && e.message == e2.message
	}

	return false
}

// Unwrap implements the Go 1.13+ Unwrap interface for use with errors.Unwrap.
func (e *Error) Unwrap() error {
	possibleError, ok := e.params[OriginalErrorParam]
	if ok {
		originalError, ok := possibleError.(error)
		if ok {
			return originalError
		}
	}

	return nil
}

// Kind returns a copy of the Error's Kind.
func (e *Error) Kind() Kind {
	return cloneKind(e.kind)
}

// WithParams adds parameters to a copy of the Error.
//
// A nil param value deletes the param key.
func (e *Error) WithParams(params Params) error {
	if len(params) == 0 {
		return e
	}

	e2 := e.clone()
	if e2.params == nil {
		e2.params = Params{}
	}

	for key, value := range params {
		if value == nil {
			delete(e2.params, key)
		} else {
			e2.params[key] = value
		}
	}

	return e2
}

// Params returns a copy of the Error's Params.
func (e *Error) Params() Params {
	return e.params.Clone()
}

// ExportRawMessage without executing the template.
func (e *Error) ExportRawMessage() string {
	return e.message
}

// Export creates a visible copy of the Error that can be used outside the erk package.
// A common use case is marshalling the error to JSON.
func (e *Error) Export() ExportedErkable {
	return &ExportedError{
		BaseExport: BaseExport{
			Kind:    GetKindString(e),
			Message: e.Error(),
			Params:  GetParams(e),
		},
	}
}

// MarshalJSON by exporting the error and then marshalling.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Export())
}

func (e *Error) clone() *Error {
	return &Error{
		kind:    e.kind,
		message: e.message,
		params:  e.Params(),
	}
}

func (e *Error) parseTemplate() (*template.Template, error) {
	t, err := template.New("").Funcs(templateFuncs(e.kind)).Parse(e.message)
	if err != nil {
		if erkstrict.IsStrictMode() {
			panic(fmt.Sprintf("Unable to parse error template:\n\tKind: %s\n\tTemplate: %s\n\tError: %v\n",
				GetKindString(e),
				e.message,
				err,
			))
		}

		return nil, err //nolint:wrapcheck // Error only used in panic
	}

	return t, nil
}
