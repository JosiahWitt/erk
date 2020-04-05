// Package erkjson allows exporting errors as JSON.
//
// To use this, first embed JSONWrapper in your project's default error kind.
// Then switch embedding erk.DefaultKind to erk.DefaultPtrKind, to enforce all kinds to be pointers.
//
// Example:
//  // Declare a default kind for your project:
//  type DefaultKind struct {
//    erk.DefaultPtrKind
//    erkjson.JSONWrapper
//  }
//
//  // Declare your kinds:
//  type ErkNotFound struct { DefaultKind }
//
//  // Create your errors:
//  var ErrSingleItemNotFound = erk.New(&ErkNotFound{}, "a single item with key '{{.key}}' was not found")
//
//  // Export your errors to JSON:
//  jsonError := erkjson.ExportError(err)
//  // jsonError is an error of type *ErkNotFound, and jsonError.Error() returns the JSON of the original error.
//
//  // The original error can be obtained with:
//  errors.Unwrap(jsonError)
//
// It is not required to embed the JSONWrapper, or use pointers to kinds.
// However, you will lose the kind as the return type.
package erkjson

import (
	"encoding/json"

	"github.com/JosiahWitt/erk"
)

// JSONWrapable must be implemented for ExportError to export the error as JSON with the kind's type.
type JSONWrapable interface {
	error
	SetJSONError(jsonError string)
	SetOriginalError(originalError error)
	IsNil() bool
}

// JSONWrapper can be embedded in error kinds.
// It is recommended to embed it in your project's default kind.
type JSONWrapper struct {
	jsonError     string
	originalError error
}

// JSONWrapper implements JSONWrapable.
var _ JSONWrapable = &JSONWrapper{}

// SetJSONError string, which is returned by Error().
func (w *JSONWrapper) SetJSONError(jsonError string) {
	w.jsonError = jsonError
}

// Error returns the JSON string.
func (w *JSONWrapper) Error() string {
	return w.jsonError
}

// SetOriginalError returned by Unwrap().
func (w *JSONWrapper) SetOriginalError(originalError error) {
	w.originalError = originalError
}

// Unwrap returns the original error.
func (w *JSONWrapper) Unwrap() error {
	return w.originalError
}

// IsNil returns true if the JSONWrapper is nil.
func (w *JSONWrapper) IsNil() bool {
	return w == nil
}

// MarshalJSON returns the same thing as Error().
func (w *JSONWrapper) MarshalJSON() ([]byte, error) {
	return []byte(w.jsonError), nil
}

// ExportError to JSON.
//
// If the error's kind implements the JSONWrapable interface, the JSON error is set on the kind.
// Otherwise, the JSON error is returned as a JSONWrapper.
func ExportError(originalError error) error {
	erkErr := erk.ToErk(originalError)

	jsonError, err := json.Marshal(erkErr)
	if err != nil {
		jsonError, _ = json.Marshal(erk.BaseExport{
			Kind:    "erk:error_is_invalid_json",
			Message: "The error cannot be wrapped as JSON: " + err.Error(),
			Params: erk.Params{
				"err": erkErr.Error(),
			},
		})
	}

	wrapper := getJSONWrapper(erkErr)
	wrapper.SetJSONError(string(jsonError))
	wrapper.SetOriginalError(originalError)
	return wrapper
}

func getJSONWrapper(err erk.Kindable) JSONWrapable {
	kind := err.Kind()
	wrapable, ok := kind.(JSONWrapable)
	if ok && !wrapable.IsNil() {
		return wrapable
	}

	return &JSONWrapper{}
}
