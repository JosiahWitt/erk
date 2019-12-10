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

// DefaultKind implements KindStringFor.
var _ KindStringFor = &DefaultKind{}

// Kindable errors that support housing an error Kind.
type Kindable interface {
	Kind() Kind
}

// KindStringFor allows a kind to override the default kind string.
type KindStringFor interface {
	KindStringFor(Kind) string
}

// IsKind checks if the error's kind is the provided kind.
func IsKind(err error, kind Kind) bool {
	return reflect.TypeOf(GetKind(err)) == reflect.TypeOf(kind)
}

// GetKind from the provided error.
func GetKind(err error) Kind {
	var k Kindable
	if errors.As(err, &k) {
		return k.Kind()
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

	// If the kind implements the KindStringFor interface, use it
	kindStrFor, ok := k.(KindStringFor)
	if ok {
		return kindStrFor.KindStringFor(k)
	}

	return (&DefaultKind{}).KindStringFor(k)
}

// KindStringFor the provided kind.
func (*DefaultKind) KindStringFor(kind Kind) string {
	t := reflect.TypeOf(kind)
	return fmt.Sprintf("%s:%s", t.PkgPath(), t.Name())
}
