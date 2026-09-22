package erg_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensuring"
	"github.com/JosiahWitt/erk/erg"
)

var _ erg.Groupable = &TestGroupable{}

type TestGroupable struct {
	errs []error
}

func (g *TestGroupable) Append(errs ...error) error {
	g.errs = append(g.errs, errs...)
	return g
}

func (g *TestGroupable) Header() error {
	return nil
}

func (g *TestGroupable) Errors() []error {
	return g.errs
}

func (g *TestGroupable) Error() string {
	return "an error"
}

func (g *TestGroupable) ErrorsString(string) string {
	return "an error"
}

func TestAppend(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with Groupable error", func(ensure ensuring.E) {
		errs := []error{errors.New("err1"), errors.New("err2")}
		g := &TestGroupable{}
		aerr := erg.Append(g, errs...)
		ensure(g.errs).Equals(errs)
		ensure(aerr.(*TestGroupable).errs).Equals(errs)
	})

	ensure.Run("with non Groupable error", func(ensure ensuring.E) {
		errs := []error{errors.New("err1"), errors.New("err2")}
		err := errors.New("not Groupable")
		aerr := erg.Append(err, errs...)
		ensure(aerr).Equals(err)
	})
}

func TestGetErrors(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with Groupable error", func(ensure ensuring.E) {
		errs := []error{errors.New("err1"), errors.New("err2")}
		g := &TestGroupable{errs: errs}
		gottenErrs := erg.GetErrors(g)
		ensure(gottenErrs).Equals(errs)
	})

	ensure.Run("with non Groupable error", func(ensure ensuring.E) {
		err := errors.New("not Groupable")
		ensure(erg.GetErrors(err)).IsNil()
	})
}

func TestAny(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with Groupable error", func(ensure ensuring.E) {
		ensure.Run("with no errors", func(ensure ensuring.E) {
			g := &TestGroupable{}
			ensure(erg.Any(g)).IsFalse()
		})

		ensure.Run("with one error", func(ensure ensuring.E) {
			errs := []error{errors.New("err1")}
			g := &TestGroupable{errs: errs}
			ensure(erg.Any(g)).IsTrue()
		})
	})

	ensure.Run("with non Groupable error", func(ensure ensuring.E) {
		err := errors.New("not Groupable")
		ensure(erg.Any(err)).IsFalse()
	})
}
