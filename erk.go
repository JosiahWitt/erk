// Package erk defines errors with kinds for Go 1.13+.
package erk

import "errors"

// Erkable errors that have Params and a Kind, and can be exported.
type Erkable interface {
	Paramable
	Kindable
	Exportable
	error
}

// ErrorIndentable allows you to specify an indent level for an error.
type ErrorIndentable interface {
	IndentError(indentLevel string) string
}

// IndentSpaces are the spaces to indent errors.
const IndentSpaces = "  "

// Wrap an error with a kind and message.
func Wrap(kind Kind, message string, err error) error {
	return WrapAs(New(kind, message), err)
}

// WrapAs wraps an error as an erk error.
func WrapAs(erkError error, err error) error {
	return WithParam(erkError, OriginalErrorParam, err)
}

// WrapWith wraps an error as an erk error with params.
//
// It is equalent to calling erk.WithParams(erk.WrapAs(erkError, err), erk.Params{}).
func WrapWith(erkError error, err error, params Params) error {
	return WithParams(WrapAs(erkError, err), params)
}

// ToErk converts an error to an erk.Erkable by wrapping it in an erk.Error.
// If it is already an erk.Erkable, it returns the error without wrapping it.
func ToErk(err error) Erkable {
	var e Erkable
	if errors.As(err, &e) {
		return e
	}

	wrappedErr := Wrap(nil, err.Error(), err).(*Error)
	wrappedErr.builtFromRegularError = err
	return wrappedErr
}
