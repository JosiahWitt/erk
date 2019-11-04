// Package erk defines errors with kinds for Go 1.13+.
package erk

import "errors"

// Erkable errors that have Params and a Kind.
type Erkable interface {
	Paramable
	Kindable
	error
}

// Wrap an error with a kind and message.
func Wrap(kind Kind, message string, err error) error {
	return WrapAs(New(kind, message), err)
}

// WrapAs wraps an error as an erkError.
func WrapAs(erkError error, err error) error {
	return WithParam(erkError, OriginalErrorParam, err)
}

// ToErk converts an error to an erk.Erkable by wrapping it in an erk.Error.
// If it is already an erk.Erkable, it returns the error without wrapping it.
func ToErk(err error) Erkable {
	var e Erkable
	if errors.As(err, &e) {
		return e
	}

	return Wrap(nil, err.Error(), err).(Erkable)
}
