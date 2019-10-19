package erk

import (
	"errors"
	"fmt"
	"reflect"
)

// Kind represents an error kind. These should be types.
//
// It is recommended to create new error kinds in each package.
// This allows erk to get the package name the error occurred in.
//
// Example:
//  package hello
//
//  type (
//    ErkJSONUnmarshalling erk.DefaultKind
//    ErkJSONMarshalling   erk.DefaultKind
//  )
//
//  ...
//
//  // Creating an error with the kind
//  err := erk.New(ErkJSONUnmarshalling, "failed to unmarshal JSON: '{{.json}}'") // Usually this would be a global error variable
//  err = erk.WithParams(err, "json", originalJSON)
//
//  ...
type Kind interface{}

// DefaultKind should be the underlying type of most Kinds.
//
// It is recommended to create new error kinds in each package.
// This allows erk to get the package name the error occurred in.
//
// Example: See Kind.
type DefaultKind struct{}

// IsKind checks if the error's kind is the provided kind.
func IsKind(err error, kind Kind) bool {
	return reflect.TypeOf(GetKind(err)) == reflect.TypeOf(kind)
}

// GetKind from the provided error.
func GetKind(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.kind
	}

	return nil
}

// GetKindString returns a string identifying what package and type of the error's kind.
//
// Example:
//  erk.GetKindString(err) // Output: "github.com/username/package:ErkYourKind"
func GetKindString(err error) string {
	k := GetKind(err)
	if k == nil {
		return ""
	}

	t := reflect.TypeOf(k)
	return fmt.Sprintf("%s:%s", t.PkgPath(), t.Name())
}
