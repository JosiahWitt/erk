package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestIsKind(t *testing.T) {
	t.Run("with equal kind", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.True(erk.IsKind(err, ErkExample{}))
	})

	t.Run("with non equal kind", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(erk.IsKind(err, ErkExample2{}), false)
	})

	t.Run("with non erk.Error not equal", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.IsKind(err, ErkExample{}), false)
	})

	t.Run("with non erk.Error equal", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.True(erk.IsKind(err, nil))
	})
}

func TestGetKind(t *testing.T) {
	t.Run("with kind", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(erk.GetKind(err), ErkExample{})
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKind(err), nil)
	})
}

func TestGetKindString(t *testing.T) {
	t.Run("with kind", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(erk.GetKindString(err), "github.com/JosiahWitt/erk_test:ErkExample")
	})

	t.Run("with non erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKindString(err), "")
	})
}
