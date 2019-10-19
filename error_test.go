package erk_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type (
	ErkExample  erk.DefaultKind
	ErkExample2 erk.DefaultKind
)

func TestNew(t *testing.T) {
	is := is.New(t)

	msg := "my message"
	err := erk.New(ErkExample{}, msg)
	is.Equal(err.Error(), msg)
	is.Equal(erk.GetParams(err), nil)
	is.Equal(erk.GetKind(err), ErkExample{})
}

func TestNewWith(t *testing.T) {
	is := is.New(t)

	msg := "my message: {{.a}}, {{.b}}!"
	err := erk.NewWith(ErkExample{}, msg, erk.Params{"a": "hello", "b": "world"})
	is.Equal(err.Error(), "my message: hello, world!")
	is.Equal(erk.GetParams(err), erk.Params{"a": "hello", "b": "world"})
	is.Equal(erk.GetKind(err), ErkExample{})
}

func TestError(t *testing.T) {
	t.Run("with invalid template", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{}}}"
		err := erk.New(ErkExample{}, msg)
		is.Equal(err.Error(), msg)
	})

	t.Run("with invalid param", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{call .a}}"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", func() { panic("just testing") })
		is.Equal(err.Error(), msg)
	})

	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}, {{.b}}!"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "hello")
		err = erk.WithParam(err, "b", "world")
		is.Equal(err.Error(), "my message: hello, world!")
	})

	t.Run("with missing params", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}, {{.b}}!"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "hello")
		is.Equal(err.Error(), "my message: hello, !")
	})
}

func TestIs(t *testing.T) {
	table := []struct {
		Name   string
		Error1 error
		Error2 error
		Equal  bool
	}{
		{
			Name:   "with two non erk.Errors",
			Error1: errors.New("one"),
			Error2: errors.New("two"),
			Equal:  false,
		},
		{
			Name:   "with the second as a non erk.Error",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: errors.New("two"),
			Equal:  false,
		},
		{
			Name:   "with both as erk.Errors with the same kind and message",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: erk.New(ErkExample{}, "my message"),
			Equal:  true,
		},
		{
			Name:   "with both as erk.Errors with the same kind and different messages",
			Error1: erk.New(ErkExample{}, "my message 1"),
			Error2: erk.New(ErkExample{}, "my message 2"),
			Equal:  false,
		},
		{
			Name:   "with both as erk.Errors with different kinds and same messages",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: erk.New(ErkExample2{}, "my message"),
			Equal:  false,
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)
			is.Equal(errors.Is(entry.Error1, entry.Error2), entry.Equal)
		})
	}
}

func TestUnwrap(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		is := is.New(t)

		errWrapped := errors.New("hey")
		err := erk.New(ErkExample{}, "my message")
		err = erk.WithParam(err, "err", errWrapped)
		is.Equal(errors.Unwrap(err), errWrapped)
	})

	t.Run("with no wrapped error", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(errors.Unwrap(err), nil)
	})
}

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

func TestToError(t *testing.T) {
	t.Run("with erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(erk.ToError(err), err)
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		msg := "the message"
		originalErr := errors.New(msg)
		wrappedErr := erk.ToError(originalErr)
		expectedErr := erk.Wrap(nil, msg, originalErr)
		is.Equal(wrappedErr, expectedErr)
		is.Equal(erk.GetKind(wrappedErr), nil)
	})
}

func TestToCopy(t *testing.T) {
	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := erk.ToCopy(err)
		is.Equal(errc.Kind, "github.com/JosiahWitt/erk_test:ErkExample")
		is.Equal(errc.Message, "my message: the world")
		is.Equal(errc.Params, erk.Params{"a": "the world"})
	})

	t.Run("returns a copy", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := erk.ToCopy(err)
		errc.Params["a"] = "123"
		is.Equal(erk.GetParams(err), erk.Params{"a": "the world"})
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		msg := "hey there"
		err := errors.New(msg)
		errc := erk.ToCopy(err)
		is.Equal(errc.Kind, "")
		is.Equal(errc.Message, msg)
		is.Equal(errc.Params, erk.Params{"err": err})
	})

	t.Run("to JSON", func(t *testing.T) {
		t.Run("with valid params", func(t *testing.T) {
			is := is.New(t)

			val := "the world"
			err := erk.New(ErkExample{}, "my message: {{.a}}")
			err = erk.WithParam(err, "a", val)
			errc := erk.ToCopy(err)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"}}`)
		})

		t.Run("with no params", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			errc := erk.ToCopy(err)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message"}`)
		})
	})
}
