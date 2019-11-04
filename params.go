package erk

import "errors"

// Paramable errors that support appending Params and getting Params.
type Paramable interface {
	WithParams(params Params) error
	Params() Params
}

// OriginalErrorParam is the param key that contains the wrapped error.
//
// This allows the original error to be used in message templates.
// Also, errors can be unwrapped, using errors.Unwrap(err).
const OriginalErrorParam = "err"

// Params are key value parameters that are usuable in the message template.
type Params map[string]interface{}

// WithParams adds parameters to an error.
//
// If err does not satisfy Paramable, the original error is returned.
// A nil param value deletes the param key.
func WithParams(err error, params Params) error {
	if len(params) == 0 {
		return err
	}

	var p Paramable
	if errors.As(err, &p) {
		return p.WithParams(params)
	}

	return err
}

// WithParam adds a parameter to an error.
//
// If err does not satisfy Paramable, the original error is returned.
// A nil param value deletes the param key.
func WithParam(err error, key string, value interface{}) error {
	return WithParams(err, Params{key: value})
}

// GetParams returns the error's parameters.
//
// If err does not satisfy Paramable, nil is returned.
func GetParams(err error) Params {
	var p Paramable
	if errors.As(err, &p) {
		return p.Params()
	}

	return nil
}
