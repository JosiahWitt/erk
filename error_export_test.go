package erk_test

import (
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
)

func TestExportedErrorErrorMessage(t *testing.T) {
	ensure := ensure.New(t)

	exportedError := erk.ExportedError{Message: "my message"}
	ensure(exportedError.ErrorMessage()).Equals("my message")
}

func TestExportedErrorErrorKind(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with nil kind", func(ensure ensurepkg.Ensure) {
		exportedError := erk.ExportedError{Kind: nil}
		ensure(exportedError.ErrorKind()).IsEmpty()
	})

	ensure.Run("with present kind", func(ensure ensurepkg.Ensure) {
		kind := "my kind"
		exportedError := erk.ExportedError{Kind: &kind}
		ensure(exportedError.ErrorKind()).Equals("my kind")
	})
}

func TestExportedErrorErrorParams(t *testing.T) {
	ensure := ensure.New(t)

	exportedError := erk.ExportedError{Params: erk.Params{"key": "value"}}
	ensure(exportedError.ErrorParams()).Equals(erk.Params{"key": "value"})
}
