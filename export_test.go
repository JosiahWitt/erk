package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestBaseExportErrorMessage(t *testing.T) {
	is := is.New(t)

	message := "my message"
	export := &erk.BaseExport{Message: message}
	is.Equal(export.ErrorMessage(), message)
}

func TestBaseExportErrorKind(t *testing.T) {
	t.Run("with non-nil kind", func(t *testing.T) {
		is := is.New(t)

		kind := "my kind"
		export := &erk.BaseExport{Kind: &kind}
		is.Equal(export.ErrorKind(), kind)
	})

	t.Run("with nil kind", func(t *testing.T) {
		is := is.New(t)

		export := &erk.BaseExport{Kind: nil}
		is.Equal(export.ErrorKind(), "")
	})
}

func TestBaseExportErrorParams(t *testing.T) {
	is := is.New(t)

	params := erk.Params{"my": "params"}
	export := &erk.BaseExport{Params: params}
	is.Equal(export.ErrorParams(), params)
}

func TestExport(t *testing.T) {
	t.Run("with erk.Error", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := erk.Export(err)
		is.Equal(errc.ErrorKind(), "github.com/JosiahWitt/erk_test:ErkExample")
		is.Equal(errc.ErrorMessage(), "my message: the world")
		is.Equal(errc.ErrorParams(), erk.Params{"a": "the world"})
	})

	t.Run("with non erk.Erkable", func(t *testing.T) {
		is := is.New(t)

		msg := "hey there"
		err := errors.New(msg)
		errc := erk.Export(err)
		is.Equal(errc.ErrorKind(), "")
		is.Equal(errc.ErrorMessage(), msg)
		is.Equal(errc.ErrorParams(), erk.Params{})
	})
}
