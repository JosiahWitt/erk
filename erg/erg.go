// Package erg allows grouping errors into an error group.
package erg

import (
	"errors"

	"github.com/JosiahWitt/erk"
)

// Groupable errors can append to and fetch the error group.
type Groupable interface {
	Header() error
	Append(errs ...error) error
	Errors() []error
}

// ExportedGroupable is an exported readonly version of Groupable.
type ExportedGroupable interface {
	GroupErrors() []erk.ExportedErkable
}

// Append to an error group.
// If groupError is not Groupable, nothing happens.
func Append(groupErr error, errs ...error) error {
	var g Groupable
	if errors.As(groupErr, &g) {
		return g.Append(errs...)
	}

	return groupErr
}

// GetErrors from an error group.
// If groupErr is not Groupable, nil is returned.
func GetErrors(groupErr error) []error {
	var g Groupable
	if errors.As(groupErr, &g) {
		return g.Errors()
	}

	return nil
}

// Any checks if there are any errors in the group.
// If groupErr is not Groupable, false is returned.
func Any(groupErr error) bool {
	var g Groupable
	if errors.As(groupErr, &g) {
		return len(g.Errors()) > 0
	}

	return false
}
