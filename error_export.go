package erk

import "errors"

// ExportedError that can be used outside the erk package.
// A common use case is marshalling the error to JSON.
type ExportedError struct {
	Kind    *string `json:"kind"`
	Type    *string `json:"type,omitempty"`
	Message string  `json:"message"`
	Params  Params  `json:"params,omitempty"`

	ErrorStack []ExportedErkable `json:"errorStack,omitempty"`
}

var _ ExportedErkable = &ExportedError{}

// ErrorMessage returns the error message.
func (e *ExportedError) ErrorMessage() string {
	return e.Message
}

// ErrorKind returns the error kind.
func (e *ExportedError) ErrorKind() string {
	if e.Kind == nil {
		return ""
	}

	return *e.Kind
}

// ErrorParams returns the error params.
func (e *ExportedError) ErrorParams() Params {
	return e.Params
}

func (e *Error) buildExportedError() *ExportedError {
	// Remove the original error from the params, since it's in the error stack
	params := GetParams(e)
	delete(params, OriginalErrorParam)

	return &ExportedError{
		Kind:       e.buildExportedKind(),
		Type:       e.buildExportedErrorType(),
		Message:    e.Error(),
		Params:     params,
		ErrorStack: nil, // This is only set at the root level by e.Export()
	}
}

func (e *Error) buildExportedKind() *string {
	if e.kind == nil {
		return nil
	}

	kindStr := e.kind.KindStringFor(e.kind)
	return &kindStr
}

func (e *Error) buildExportedErrorType() *string {
	if e.builtFromRegularError == nil {
		return nil
	}

	typeStr := buildDefaultKindString(e.builtFromRegularError)
	return &typeStr
}

func (e *Error) buildErrorStack() []ExportedErkable {
	errs := []ExportedErkable{}

	for currentErr := errors.Unwrap(e); currentErr != nil; currentErr = errors.Unwrap(currentErr) {
		// If we converted a regular error to an erk error, don't include the error itself in the error stack
		//nolint:err113 // Our intention is to directly compare the error, not unwrap and compare other errors
		if e.builtFromRegularError != nil && e.builtFromRegularError == currentErr {
			continue
		}

		exportedErkErr := buildErrorStackEntry(currentErr)
		errs = append(errs, exportedErkErr)
	}

	return errs
}

func buildErrorStackEntry(currentErr error) ExportedErkable {
	currentErkableErr := ToErk(currentErr)
	currentErkErr, ok := currentErkableErr.(*Error)
	if !ok {
		return currentErkableErr.Export()
	}

	return currentErkErr.buildExportedError()
}
