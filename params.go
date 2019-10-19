package erk

import "errors"

// OriginalErrorParam is the param key that contains the wrapped error.
//
// This allows the original error to be used in message templates.
// Also, errors can be unwrapped, using errors.Unwrap(err).
const OriginalErrorParam = "err"

// Params are key value parameters that are usuable in the message template.
type Params map[string]interface{}

// WithParams on a copy of the error.
//
// If err is not an erk.Error, it is converted to one first by calling erk.ToError.
// A nil param value deletes the param key.
func WithParams(err error, params Params) error {
	if len(params) == 0 {
		return err
	}

	e := clone(err)
	if e.params == nil {
		e.params = Params{}
	}

	for key, value := range params {
		if value == nil {
			delete(e.params, key)
		} else {
			e.params[key] = value
		}
	}

	return e
}

// WithParam on a copy of the error.
//
// If err is not an erk.Error, it is converted to one first by calling erk.ToError.
// A nil param value deletes the param key.
func WithParam(err error, key string, value interface{}) error {
	return WithParams(err, Params{key: value})
}

// GetParams returns a copy of the error's parameters.
//
// If no parameters have been set, or err is not an erk.Error, nil is returned.
func GetParams(err error) Params {
	var e *Error
	if errors.As(err, &e) {
		if e.params == nil {
			return nil
		}

		paramsCopy := Params{}
		for k, v := range e.params {
			paramsCopy[k] = v
		}

		return paramsCopy
	}

	return nil
}
