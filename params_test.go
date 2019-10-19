package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestWithParams(t *testing.T) {
	t.Run("with erk.Error", func(t *testing.T) {
		t.Run("with nil params, setting nil params", func(t *testing.T) {
			is := is.New(t)

			err1 := erk.New(ErkExample{}, "my message")
			err2 := erk.WithParams(err1, nil)
			is.Equal(err2, err1)
			is.Equal(erk.GetParams(err2), nil)
		})

		t.Run("with nil params, setting two params", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			err = erk.WithParams(err, erk.Params{"a": "hello", "b": "world"})
			is.Equal(erk.GetParams(err), erk.Params{"a": "hello", "b": "world"})
		})

		t.Run("with present params, setting nil params", func(t *testing.T) {
			is := is.New(t)

			err1 := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err2 := erk.WithParams(err1, nil)
			is.Equal(err2, err1)
			is.Equal(erk.GetParams(err2), erk.Params{"0": "hey", "1": "there"})
		})

		t.Run("with present params, setting two params", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err = erk.WithParams(err, erk.Params{"a": "hello", "b": "world"})
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
		})

		t.Run("with present params, deleting one param", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err = erk.WithParams(err, erk.Params{"a": "hello", "b": "world", "1": nil})
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "a": "hello", "b": "world"})
		})
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		t.Run("setting nil params", func(t *testing.T) {
			is := is.New(t)

			err1 := errors.New("hi")
			err2 := erk.WithParams(err1, nil)
			is.Equal(err2, err1)
			is.Equal(erk.GetParams(err2), nil)
		})

		t.Run("setting two params", func(t *testing.T) {
			is := is.New(t)

			err1 := errors.New("hi")
			err2 := erk.WithParams(err1, erk.Params{"a": "hello", "b": "world"})
			is.Equal(erk.GetParams(err2), erk.Params{"a": "hello", "b": "world", "err": err1})
		})
	})
}

func TestWithParam(t *testing.T) {
	t.Run("with erk.Error", func(t *testing.T) {
		t.Run("with nil params", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			err = erk.WithParam(err, "a", "hello")
			err = erk.WithParam(err, "b", "world")
			is.Equal(erk.GetParams(err), erk.Params{"a": "hello", "b": "world"})
		})

		t.Run("with present params", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err = erk.WithParam(err, "a", "hello")
			err = erk.WithParam(err, "b", "world")
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
		})
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		err1 := errors.New("hi")
		err2 := erk.WithParam(err1, "a", "hello")
		err2 = erk.WithParam(err2, "b", "world")
		is.Equal(erk.GetParams(err2), erk.Params{"a": "hello", "b": "world", "err": err1})
	})
}

func TestGetParams(t *testing.T) {
	t.Run("with erk.Error", func(t *testing.T) {
		t.Run("returns parameters", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there"})
		})

		t.Run("returns a copy of the parameters", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			params := erk.GetParams(err)
			params["0"] = "changed"
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there"})
		})
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("hi")
		is.Equal(erk.GetParams(err), nil)
	})
}
