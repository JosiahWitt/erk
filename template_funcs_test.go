package erk_test

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

func TestTemplateFuncs(t *testing.T) {
	t.Run("with no TemplateFuncsFor function", func(t *testing.T) {
		is := is.New(t)

		type TestType string

		msg := "{{.a}} is {{type .a}}"
		err := erk.New(ErkSimple{}, msg)
		err = erk.WithParam(err, "a", TestType("hello"))
		is.Equal(err.Error(), "hello is erk_test.TestType")
	})

	t.Run("with overridden TemplateFuncsFor function", func(t *testing.T) {
		is := is.New(t)

		type TestType string

		msg := "{{.a}} is {{fancyType .a}}"
		err := erk.New(ErkOverriddenTemplateFuncs{}, msg)
		err = erk.WithParam(err, "a", TestType("hello"))
		is.Equal(err.Error(), "hello is 'type from overridden_funcs: erk_test.TestType'")
	})

	t.Run("with default kind", func(t *testing.T) {
		t.Run("when printing type of param", func(t *testing.T) {
			is := is.New(t)

			type TestType string

			msg := "{{.a}} is {{type .a}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", TestType("hello"))
			is.Equal(err.Error(), "hello is erk_test.TestType")
		})

		t.Run("when inspecting complex param", func(t *testing.T) {
			is := is.New(t)

			type param struct {
				Msg string
				Map map[string]string
			}

			msg := "my message: {{inspect .a}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", param{Msg: "hey", Map: map[string]string{"key": "value"}})
			is.Equal(err.Error(), "my message: {Msg:hey Map:map[key:value]}")
		})
	})
}

type ErkSimple struct{}

var _ erk.Kind = ErkSimple{}

func (ErkSimple) KindStringFor(erk.Kind) string {
	return "erk_simple"
}

type ErkOverriddenTemplateFuncs struct{ erk.DefaultKind }

func (k ErkOverriddenTemplateFuncs) TemplateFuncsFor(k2 erk.Kind) template.FuncMap {
	funcMap := k.DefaultKind.TemplateFuncsFor(k2)
	funcMap["fancyType"] = func(v interface{}) string {
		return fmt.Sprintf("'type from %s: %T'", k2.KindStringFor(k2), v)
	}

	return funcMap
}

func (ErkOverriddenTemplateFuncs) KindStringFor(erk.Kind) string {
	return "overridden_funcs"
}
