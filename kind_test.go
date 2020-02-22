package erk_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type TestKindable struct {
	kind erk.Kind
}

func (k *TestKindable) Kind() erk.Kind {
	return k.kind
}

func (k *TestKindable) Error() string {
	return fmt.Sprintf("%T", k.kind)
}

type TestKindStringFor struct{ erk.DefaultKind }

func (TestKindStringFor) KindStringFor(k erk.Kind) string {
	return "my_kind"
}

func TestIsKind(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		t.Run("with equal kind", func(t *testing.T) {
			is := is.New(t)

			err := &TestKindable{kind: ErkExample{}}
			is.True(erk.IsKind(err, ErkExample{}))
		})

		t.Run("with non equal kind", func(t *testing.T) {
			is := is.New(t)

			err := &TestKindable{kind: ErkExample{}}
			is.Equal(erk.IsKind(err, ErkExample2{}), false)
		})
	})

	t.Run("with erk.Error", func(t *testing.T) {
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
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		t.Run("with not equal kind", func(t *testing.T) {
			is := is.New(t)

			err := errors.New("abc")
			is.Equal(erk.IsKind(err, ErkExample{}), false)
		})

		t.Run("with equal kind", func(t *testing.T) {
			is := is.New(t)

			err := errors.New("abc")
			is.True(erk.IsKind(err, nil))
		})
	})
}

func TestGetKind(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: ErkExample{}}
		is.Equal(erk.GetKind(err), ErkExample{})
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKind(err), nil)
	})
}

func TestGetKindString(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: ErkExample{}}
		is.Equal(erk.GetKindString(err), "github.com/JosiahWitt/erk_test:ErkExample")
	})

	t.Run("with erk.KindStringFor", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: TestKindStringFor{}}
		is.Equal(erk.GetKindString(err), "my_kind")
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKindString(err), "")
	})
}

func TestKindTemplateFuncsFor(t *testing.T) {
	is := is.New(t)

	funcMap := erk.DefaultKind{}.TemplateFuncsFor(erk.DefaultKind{})
	funcMap["abc"] = func() string { return "hey" }

	funcMap2 := erk.DefaultKind{}.TemplateFuncsFor(erk.DefaultKind{})
	_, ok := funcMap2["abc"]
	is.True(!ok) // Returned func map should be a copy
}
