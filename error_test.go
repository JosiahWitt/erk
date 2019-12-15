package erk_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type (
	ErkExample  struct { erk.DefaultKind }
	ErkExample2 struct { erk.DefaultKind }
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
		is.Equal(err.Error(), "my message: hello, <no value>!")
	})

	t.Run("with param with quotes", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "'quoted'")
		is.Equal(err.Error(), "my message: 'quoted'")
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

func TestErrorKind(t *testing.T) {
	is := is.New(t)

	err := erk.New(ErkExample{}, "my message")
	is.Equal(err.(*erk.Error).Kind(), ErkExample{})
}

func TestErrorWithParams(t *testing.T) {
	t.Run("with nil params, setting nil params", func(t *testing.T) {
		is := is.New(t)

		err1 := erk.New(ErkExample{}, "my message")
		err2 := err1.(*erk.Error).WithParams(nil)
		is.Equal(err2, err1)
		is.Equal(err2.(*erk.Error).Params(), nil)
	})

	t.Run("with nil params, setting two params", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"a": "hello", "b": "world"})
	})

	t.Run("with present params, setting nil params", func(t *testing.T) {
		is := is.New(t)

		err1 := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err2 := err1.(*erk.Error).WithParams(nil)
		is.Equal(err2, err1)
		is.Equal(err2.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("with present params, setting two params", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
	})

	t.Run("with present params, deleting one param", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world", "1": nil})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "a": "hello", "b": "world"})
	})
}

func TestErrorParams(t *testing.T) {
	t.Run("returns parameters", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("returns a copy of the parameters", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		params := err.(*erk.Error).Params()
		params["0"] = "changed"
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})
}

func TestErrorExport(t *testing.T) {
	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)
		is.Equal(errc.Kind, "github.com/JosiahWitt/erk_test:ErkExample")
		is.Equal(errc.Message, "my message: the world")
		is.Equal(errc.Params, erk.Params{"a": "the world"})
	})

	t.Run("returns a copy", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)
		errc.Params["a"] = "123"
		is.Equal(erk.GetParams(err), erk.Params{"a": "the world"})
	})

	t.Run("to JSON", func(t *testing.T) {
		t.Run("with valid params", func(t *testing.T) {
			is := is.New(t)

			val := "the world"
			err := erk.New(ErkExample{}, "my message: {{.a}}")
			err = erk.WithParam(err, "a", val)
			errc := err.(*erk.Error).Export().(*erk.ExportedError)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"}}`)
		})

		t.Run("with no params", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			errc := err.(*erk.Error).Export().(*erk.ExportedError)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message"}`)
		})
	})
}
