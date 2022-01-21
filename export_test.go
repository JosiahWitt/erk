package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
)

func TestBaseExportErrorMessage(t *testing.T) {
	ensure := ensure.New(t)

	message := "my message"
	export := &erk.BaseExport{Message: message}
	ensure(export.ErrorMessage()).Equals(message)
}

func TestBaseExportErrorKind(t *testing.T) {
	ensure := ensure.New(t)

	export := &erk.BaseExport{Kind: "error kind"}
	ensure(export.ErrorKind()).Equals("error kind")
}

func TestBaseExportErrorParams(t *testing.T) {
	ensure := ensure.New(t)

	params := erk.Params{"my": "params"}
	export := &erk.BaseExport{Params: params}
	ensure(export.ErrorParams()).Equals(params)
}

func TestExport(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Error", func(ensure ensurepkg.Ensure) {
		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := erk.Export(err)
		ensure(errc.ErrorKind()).Equals("github.com/JosiahWitt/erk_test:ErkExample")
		ensure(errc.ErrorMessage()).Equals("my message: the world")
		ensure(errc.ErrorParams()).Equals(erk.Params{"a": "the world"})
	})

	ensure.Run("with non erk.Erkable", func(ensure ensurepkg.Ensure) {
		msg := "hey there"
		err := errors.New(msg)
		errc := erk.Export(err)
		ensure(errc.ErrorKind()).IsEmpty()
		ensure(errc.ErrorMessage()).Equals(msg)
		ensure(errc.ErrorParams()).Equals(erk.Params{})
	})
}
