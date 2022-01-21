package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
)

func TestWrap(t *testing.T) {
	ensure := ensure.New(t)

	errWrapped := errors.New("hey")
	msg := "my message"
	err := erk.Wrap(ErkExample{}, msg, errWrapped)

	ensure(err.Error()).Equals(msg)
	ensure(erk.GetParams(err)).Equals(erk.Params{"err": errWrapped})
	ensure(erk.GetKind(err)).Equals(ErkExample{})
	ensure(errors.Unwrap(err)).Equals(errWrapped)
}

func TestWrapAs(t *testing.T) {
	ensure := ensure.New(t)

	errWrapped := errors.New("hey")
	msg := "my message"
	errWrapper := erk.New(ErkExample{}, msg)
	err := erk.WrapAs(errWrapper, errWrapped)

	ensure(err.Error()).Equals(msg)
	ensure(erk.GetParams(err)).Equals(erk.Params{"err": errWrapped})
	ensure(erk.GetKind(err)).Equals(ErkExample{})
	ensure(errors.Unwrap(err)).Equals(errWrapped)
}

func TestWrapWith(t *testing.T) {
	ensure := ensure.New(t)

	errWrapped := errors.New("hey")
	msg := "my message: {{.a}}"
	errWrapper := erk.New(ErkExample{}, msg)
	err := erk.WrapWith(errWrapper, errWrapped, erk.Params{"a": "hello"})

	ensure(err.Error()).Equals("my message: hello")
	ensure(erk.GetParams(err)).Equals(erk.Params{"err": errWrapped, "a": "hello"})
	ensure(erk.GetKind(err)).Equals(ErkExample{})
	ensure(errors.Unwrap(err)).Equals(errWrapped)
}

func TestToErk(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Erkable", func(ensure ensurepkg.Ensure) {
		err := erk.New(ErkExample{}, "my message")
		ensure(erk.ToErk(err)).Equals(err)
	})

	ensure.Run("with non erk.Erkable", func(ensure ensurepkg.Ensure) {
		msg := "the message"
		originalErr := errors.New(msg)
		wrappedErr := erk.ToErk(originalErr)
		ensure(erk.GetKind(wrappedErr)).IsNil()
		ensure(wrappedErr.Error()).Equals(originalErr.Error())
		ensure(erk.GetParams(wrappedErr)).Equals(erk.Params{erk.OriginalErrorParam: originalErr})
	})
}
