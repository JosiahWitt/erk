// Package erkmock allows creating erk errors to be returned from mocked interfaces.
// Without using this package, it's possible to get false positive strict mode errors.
//
// Example:
//
//	someMockedFunction.Returns(erkmock.From(store.ErrItemNotFound))
package erkmock

import "github.com/JosiahWitt/erk"

// From a given erk error, create a mock error.
// This will panic if the provided error does not satisfy erk.Erkable.
func From(err error) error {
	erkable, ok := err.(erk.Erkable)
	if !ok {
		panic("erkmock.From only supports mocking erk.Erkable errors")
	}

	mockError := For(erkable.Kind())
	mockError.(*Mock).SetMessage(erkable.ExportRawMessage()) //nolint:forcetypeassert // We know this is a Mock
	return mockError
}
