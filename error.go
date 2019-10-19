// Package erk defines errors with kinds for Go 1.13+.
package erk

import (
	"bytes"
	"errors"
	"html/template"
)

// Error stores details about an error with kinds and a message template.
type Error struct {
	kind    Kind
	message string
	params  Params
}

// ErrorCopy represents a copy of the error.
// A common use case is marshalling to JSON.
type ErrorCopy struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Params  Params `json:"params,omitempty"`
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

// Wrap an error with a kind and message.
func Wrap(kind Kind, message string, err error) error {
	return WrapAs(New(kind, message), err)
}

// WrapAs wraps an error as an erkError.
func WrapAs(erkError error, err error) error {
	return WithParam(erkError, OriginalErrorParam, err)
}

// ToError converts an error to an erk.Error by wrapping it.
// If it is already an erk.Error, it returns the error without wrapping it.
func ToError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	return Wrap(nil, err.Error(), err).(*Error)
}

// ToCopy creates a visible copy of the error that can be used outside the erk package.
// A common use case is marshalling the error to JSON.
// If err is not an erk.Error, it is wrapped first.
func ToCopy(err error) *ErrorCopy {
	e := ToError(err)

	return &ErrorCopy{
		Kind:    GetKindString(e),
		Message: e.Error(),
		Params:  GetParams(e),
	}
}

func clone(err error) *Error {
	e := ToError(err)

	return &Error{
		kind:    e.kind,
		message: e.message,
		params:  GetParams(e),
	}
}
