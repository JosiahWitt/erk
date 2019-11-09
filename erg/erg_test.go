package erg_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk/erg"
	"github.com/matryer/is"
)

var _ erg.Groupable = &TestGroupable{}

type TestGroupable struct {
	errs []error
}

func (g *TestGroupable) Append(errs ...error) error {
	g.errs = append(g.errs, errs...)
	return g
}

func (g *TestGroupable) Errors() []error {
	return g.errs
}

func (g *TestGroupable) Error() string {
	return "an error"
}

func TestAppend(t *testing.T) {
	t.Run("with Groupable error", func(t *testing.T) {
		is := is.New(t)

		errs := []error{errors.New("err1"), errors.New("err2")}
		g := &TestGroupable{}
		aerr := erg.Append(g, errs...)
		is.Equal(g.errs, errs)
		is.Equal(aerr.(*TestGroupable).errs, errs)
	})

	t.Run("with non Groupable error", func(t *testing.T) {
		is := is.New(t)

		errs := []error{errors.New("err1"), errors.New("err2")}
		err := errors.New("not Groupable")
		aerr := erg.Append(err, errs...)
		is.Equal(aerr, err)
	})
}

func TestGetErrors(t *testing.T) {
	t.Run("with Groupable error", func(t *testing.T) {
		is := is.New(t)

		errs := []error{errors.New("err1"), errors.New("err2")}
		g := &TestGroupable{errs: errs}
		gottenErrs := erg.GetErrors(g)
		is.Equal(gottenErrs, errs)
	})

	t.Run("with non Groupable error", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("not Groupable")
		is.Equal(erg.GetErrors(err), nil)
	})
}

func TestAny(t *testing.T) {
	t.Run("with Groupable error", func(t *testing.T) {
		t.Run("with no errors", func(t *testing.T) {
			is := is.New(t)

			g := &TestGroupable{}
			is.Equal(erg.Any(g), false)
		})

		t.Run("with one error", func(t *testing.T) {
			is := is.New(t)

			errs := []error{errors.New("err1")}
			g := &TestGroupable{errs: errs}
			is.Equal(erg.Any(g), true)
		})
	})

	t.Run("with non Groupable error", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("not Groupable")
		is.Equal(erg.Any(err), false)
	})
}
