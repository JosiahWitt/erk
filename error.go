package erk

import (
	"bytes"
	"errors"
	"html/template"
)

// Error satisfies the Erkable interface.
var _ Erkable = &Error{}

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
	return &Error{
		kind:    kind,
		message: message,
	}
}

// NewWith creates an error with a kind, message, and params.
func NewWith(kind Kind, message string, params Params) error {
	return &Error{
		kind:    kind,
		message: message,
		params:  params,
	}
}

// Error processes the message template with the provided params.
func (e *Error) Error() string {
	t, err := template.New("").Parse(e.message)
	if err != nil {
		return e.message
	}

	var filledMessage bytes.Buffer
	err = t.Execute(&filledMessage, e.params)
	if err != nil {
		return e.message
	}

	return filledMessage.String()
}

// Is implements the Go 1.13+ Is interface for use with errors.Is.
func (e *Error) Is(err error) bool {
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

// Kind of the Error.
// See Kind for more details.
func (e *Error) Kind() Kind {
	return e.kind
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
	if e.params == nil {
		return nil
	}

	paramsCopy := Params{}
	for k, v := range e.params {
		paramsCopy[k] = v
	}

	return paramsCopy
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

func (e *Error) clone() *Error {
	return &Error{
		kind:    e.kind,
		message: e.message,
		params:  e.Params(),
	}
}
