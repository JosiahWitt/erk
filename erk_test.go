package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestWrap(t *testing.T) {
	is := is.New(t)

	errWrapped := errors.New("hey")
	msg := "my message"
	err := erk.Wrap(ErkExample{}, msg, errWrapped)

	is.Equal(err.Error(), msg)
	is.Equal(erk.GetParams(err), erk.Params{"err": errWrapped})
	is.Equal(erk.GetKind(err), ErkExample{})
	is.Equal(errors.Unwrap(err), errWrapped)
}

func TestWrapAs(t *testing.T) {
	is := is.New(t)

	errWrapped := errors.New("hey")
	msg := "my message"
	errWrapper := erk.New(ErkExample{}, msg)
	err := erk.WrapAs(errWrapper, errWrapped)

	is.Equal(err.Error(), msg)
	is.Equal(erk.GetParams(err), erk.Params{"err": errWrapped})
	is.Equal(erk.GetKind(err), ErkExample{})
	is.Equal(errors.Unwrap(err), errWrapped)
}

func TestWrapWith(t *testing.T) {
	is := is.New(t)

	errWrapped := errors.New("hey")
	msg := "my message: {{.a}}"
	errWrapper := erk.New(ErkExample{}, msg)
	err := erk.WrapWith(errWrapper, errWrapped, erk.Params{"a": "hello"})

	is.Equal(err.Error(), "my message: hello")
	is.Equal(erk.GetParams(err), erk.Params{"err": errWrapped, "a": "hello"})
	is.Equal(erk.GetKind(err), ErkExample{})
	is.Equal(errors.Unwrap(err), errWrapped)
}

func TestToErk(t *testing.T) {
	t.Run("with erk.Erkable", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(erk.ToErk(err), err)
	})

	t.Run("with non erk.Erkable", func(t *testing.T) {
		is := is.New(t)

		msg := "the message"
		originalErr := errors.New(msg)
		wrappedErr := erk.ToErk(originalErr)
		is.Equal(erk.GetKind(wrappedErr), nil)
		is.Equal(wrappedErr.Error(), originalErr.Error())
		is.Equal(erk.GetParams(wrappedErr), erk.Params{erk.OriginalErrorParam: originalErr})
	})
}
