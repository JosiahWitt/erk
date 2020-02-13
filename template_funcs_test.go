package erk_test

import (
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestTemplateFuncs(t *testing.T) {
	t.Run("when printing type of param", func(t *testing.T) {
		is := is.New(t)

		type TestType string

		msg := "{{.a}} is {{type .a}}"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", TestType("hello"))
		is.Equal(err.Error(), "hello is erk_test.TestType")
	})
}
