// Package erk provides error helpers for Go 1.13+.
package erk

import (
	"errors"
	"reflect"
)

var stringErrorType = reflect.TypeOf(errors.New(""))

// IsA compares the source error with the target error.
// If errors.Is is true, then IsA returns true.
// Otherwise, their types are compared.
//
// Note: IsA returns false if target was created using errors.New, as this seems dangerous.
// If you want to check if an error was created using errors.New, use the IsAStringError function.
func IsA(source, target error) bool {
	if errors.Is(source, target) {
		return true
	}

	sourceType := reflect.TypeOf(source)
	targetType := reflect.TypeOf(target)
	return targetType != stringErrorType && sourceType == targetType
}

// IsAStringError returns true if err was created using errors.New and false otherwise.
func IsAStringError(err error) bool {
	return reflect.TypeOf(err) == stringErrorType
}
