package erk_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkstrict"
	"github.com/matryer/is"
)

//nolint:gochecknoinits // Used to enforce false strict mode
func init() {
	erkstrict.SetStrictMode(false)
}

type (
	ErkExample  struct{ erk.DefaultKind }
	ErkExample2 struct{ erk.DefaultKind }
)

func TestNew(t *testing.T) {
	testNew(t, func(kind erk.Kind, message string, params erk.Params) error {
		err := erk.New(kind, message)
		return erk.WithParams(err, params)
	})
}

func TestNewWith(t *testing.T) {
	testNew(t, erk.NewWith)
}

func testNew(t *testing.T, create func(kind erk.Kind, message string, params erk.Params) error) {
	validTemplate := func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{inspect .a}}, {{.b}}!"
		err := create(ErkExample{}, msg, erk.Params{"a": "hello", "b": "world"})
		is.Equal(err.Error(), "my message: hello, world!")
		is.Equal(erk.GetParams(err), erk.Params{"a": "hello", "b": "world"})
		is.Equal(erk.GetKind(err), ErkExample{})
	}

	t.Run("no strict mode", func(t *testing.T) {
		t.Run("valid template", validTemplate)

		t.Run("invalid template", func(t *testing.T) {
			is := is.New(t)

			msg := "my message: {{}}!"
			err := create(ErkExample{}, msg, erk.Params{})
			is.Equal(err.Error(), "my message: {{}}!")
		})
	})

	t.Run("strict mode", func(t *testing.T) {
		withStrictMode(true, func() {
			t.Run("valid template", validTemplate)

			t.Run("invalid template", func(t *testing.T) {
				is := is.New(t)

				defer func() {
					if res := recover(); res != nil {
						str, ok := res.(string)
						is.True(ok)

						is.True(strings.Contains(str, "Unable to parse error template"))
						is.True(strings.Contains(str, "Template: my message {{}}}"))
						is.True(strings.Contains(str, "Error: "))
					}
				}()

				msg := "my message {{}}}"
				create(ErkExample{}, msg, nil) //nolint:errcheck // Used to trigger panic
				is.Fail()                      // Expected panic
			})
		})
	})
}

func TestError(t *testing.T) {
	t.Run("with invalid template", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{}}}"
		err := erk.New(ErkExample{}, msg)
		is.Equal(err.Error(), msg)
	})

	t.Run("with invalid param", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{call .a}}"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", func() { panic("just testing") })
		is.Equal(err.Error(), msg)
	})

	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}, {{.b}}!"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "hello")
		err = erk.WithParam(err, "b", "world")
		is.Equal(err.Error(), "my message: hello, world!")
	})

	t.Run("with missing params", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}, {{.b}}!"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "hello")
		is.Equal(err.Error(), "my message: hello, <no value>!")
	})

	t.Run("with param with quotes", func(t *testing.T) {
		is := is.New(t)

		msg := "my message: {{.a}}"
		err := erk.New(ErkExample{}, msg)
		err = erk.WithParam(err, "a", "'quoted'")
		is.Equal(err.Error(), "my message: 'quoted'")
	})

	t.Run("with wrapped error", func(t *testing.T) {
		t.Run("with no newlines", func(t *testing.T) {
			is := is.New(t)

			wrappedErr := errors.New("see! there are no newlines; this one (\\n) is escaped :)")
			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WrapAs(err, wrappedErr)
			is.Equal(err.Error(), "my message: see! there are no newlines; this one (\\n) is escaped :)")
		})

		t.Run("with newlines", func(t *testing.T) {
			is := is.New(t)

			wrappedErr := errors.New("a group:\n - item one\n - item two")
			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WrapAs(err, wrappedErr)
			is.Equal(err.Error(), "my message: \n  a group:\n   - item one\n   - item two")
		})

		t.Run("with newlines two layers deep", func(t *testing.T) {
			is := is.New(t)

			wrappedErr := errors.New("a group:\n - item one\n - item two")
			msgNested := "my message nested: {{.err}}"
			errNested := erk.New(ErkExample{}, msgNested)
			errNested = erk.WrapAs(errNested, wrappedErr)

			msg := "my message 1: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WrapAs(err, errNested)

			is.Equal(err.Error(), "my message 1: my message nested: \n  a group:\n   - item one\n   - item two")
		})

		// For now we don't worry about the case when an erk message contains newlines.
		// We can revisit this later if there is a valid use case.
		t.Run("with newlines in wrapped erk error", func(t *testing.T) {
			is := is.New(t)

			wrappedErr := erk.NewWith(ErkExample{}, "a group:\n - item one\n - item {{.twoName}}", erk.Params{"twoName": "two"})
			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WrapAs(err, wrappedErr)
			is.Equal(err.Error(), "my message: a group:\n - item one\n - item two")
		})

		t.Run("that wasn't wrapped", func(t *testing.T) {
			is := is.New(t)

			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			is.Equal(err.Error(), "my message: <no value>")
		})

		t.Run("with err param that isn't an error but contains newlines", func(t *testing.T) {
			is := is.New(t)

			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "err", "hey \nnewline")
			is.Equal(err.Error(), "my message: hey \nnewline")
		})
	})
}

func TestErrorStrictMode(t *testing.T) {
	t.Run("with erk strict disabled", func(t *testing.T) {
		t.Run("with invalid template", func(t *testing.T) {
			is := is.New(t)

			msg := "my message {{}}}"
			err := erk.New(ErkExample{}, msg)
			is.Equal(err.Error(), msg)
		})

		t.Run("with invalid param", func(t *testing.T) {
			is := is.New(t)

			msg := "my message {{call .a}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", func() { panic("just testing") })
			is.Equal(err.Error(), msg)
		})

		t.Run("with missing params", func(t *testing.T) {
			is := is.New(t)

			msg := "my message: {{.a}}, {{.b}}!"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", "hello")
			is.Equal(err.Error(), "my message: hello, <no value>!")
		})
	})

	t.Run("with erk strict enabled", func(t *testing.T) {
		t.Run("with invalid template", func(t *testing.T) {
			is := is.New(t)

			defer func() {
				if res := recover(); res != nil {
					str, ok := res.(string)
					is.True(ok)

					is.True(strings.Contains(str, "Unable to parse error template"))
					is.True(strings.Contains(str, "Template: my message {{}}}"))
					is.True(strings.Contains(str, "Error: "))
				}
			}()

			msg := "my message {{}}}"
			err := erk.New(ErkExample{}, msg)

			withStrictMode(true, func() { err.Error() }) //nolint:govet // Used to trigger panic
			is.Fail()                                    // Expected panic
		})

		t.Run("with invalid param", func(t *testing.T) {
			is := is.New(t)

			defer func() {
				if res := recover(); res != nil {
					str, ok := res.(string)
					is.True(ok)

					is.True(strings.Contains(str, "Unable to execute error template"))
					is.True(strings.Contains(str, "Kind: github.com/JosiahWitt/erk_test:ErkExample"))
					is.True(strings.Contains(str, "Template: my message {{call .a}}"))
					is.True(strings.Contains(str, "Params: map[a:"))
					is.True(strings.Contains(str, "Error: "))
				}
			}()

			msg := "my message {{call .a}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", func() { panic("just testing") })
			withStrictMode(true, func() { err.Error() }) //nolint:govet // Used to trigger panic
			is.Fail()                                    // Expected panic
		})

		t.Run("with missing params", func(t *testing.T) {
			is := is.New(t)

			defer func() {
				if res := recover(); res != nil {
					str, ok := res.(string)
					is.True(ok)

					is.True(strings.Contains(str, "Unable to execute error template"))
					is.True(strings.Contains(str, "Kind: github.com/JosiahWitt/erk_test:ErkExample"))
					is.True(strings.Contains(str, "Template: my message: {{.a}}, {{.b}}!"))
					is.True(strings.Contains(str, "Params: map[a:hello]"))
					is.True(strings.Contains(str, "Error: "))
				}
			}()

			msg := "my message: {{.a}}, {{.b}}!"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", "hello")
			withStrictMode(true, func() { err.Error() }) //nolint:govet // Used to trigger panic
			is.Fail()                                    // Expected panic
		})
	})
}

func TestIs(t *testing.T) {
	table := []struct {
		Name       string
		Error1     error
		Error2     error
		Equal      bool
		Panic      bool
		StrictMode bool
	}{
		{
			Name:   "with two non erk.Errors",
			Error1: errors.New("one"),
			Error2: errors.New("two"),
			Equal:  false,
		},
		{
			Name:   "with the second as a non erk.Error",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: errors.New("two"),
			Equal:  false,
		},
		{
			Name:   "with both as erk.Errors with the same kind and message",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: erk.New(ErkExample{}, "my message"),
			Equal:  true,
		},
		{
			Name:   "with both as erk.Errors with the same kind and different messages",
			Error1: erk.New(ErkExample{}, "my message 1"),
			Error2: erk.New(ErkExample{}, "my message 2"),
			Equal:  false,
		},
		{
			Name:   "with both as erk.Errors with different kinds and same messages",
			Error1: erk.New(ErkExample{}, "my message"),
			Error2: erk.New(ErkExample2{}, "my message"),
			Equal:  false,
		},
		{
			Name:       "with invalid template and no strict mode and equal",
			Error1:     erk.New(ErkExample{}, "my message {{}}"),
			Error2:     erk.New(ErkExample{}, "my message {{}}"),
			Equal:      true,
			StrictMode: false,
		},
		{
			Name:       "with invalid template and strict mode and equal",
			Error1:     erk.New(ErkExample{}, "my message {{}}"),
			Error2:     erk.New(ErkExample{}, "my message {{}}"),
			Equal:      true,
			StrictMode: true,
			Panic:      true,
		},
		{
			Name:       "with invalid template and strict mode and not equal",
			Error1:     erk.New(ErkExample{}, "my message {{}}"),
			Error2:     errors.New("another error"),
			Equal:      false,
			StrictMode: true,
			Panic:      true,
		},
	}

	for _, entry := range table {
		entry := entry // Pin range variable

		t.Run(entry.Name, func(t *testing.T) {
			withStrictMode(entry.StrictMode, func() {
				is := is.New(t)

				defer func() {
					if res := recover(); res != nil {
						is.True(entry.Panic)
					}
				}()

				is.Equal(errors.Is(entry.Error1, entry.Error2), entry.Equal)
				is.Equal(false, entry.Panic)
			})
		})
	}
}

func TestUnwrap(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		is := is.New(t)

		errWrapped := errors.New("hey")
		err := erk.New(ErkExample{}, "my message")
		err = erk.WithParam(err, "err", errWrapped)
		is.Equal(errors.Unwrap(err), errWrapped)
	})

	t.Run("with no wrapped error", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(errors.Unwrap(err), nil)
	})
}

func TestErrorKind(t *testing.T) {
	t.Run("simple clone", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		is.Equal(err.(*erk.Error).Kind(), ErkExample{})
	})

	t.Run("kind as a pointer is cloned", func(t *testing.T) {
		is := is.New(t)

		originalKind := &KindWithField{Field: "hey"}
		expectedKind := &KindWithField{Field: "hey"}
		err := erk.New(originalKind, "my message")

		kindCopy, ok := err.(*erk.Error).Kind().(*KindWithField)
		is.True(ok)
		is.Equal(kindCopy, expectedKind)

		kindCopy.Field = "something else"
		is.Equal(originalKind, expectedKind) // It should not modify the original kind
	})

	t.Run("kind that doesn't implement the CloneKind function is cloned", func(t *testing.T) {
		is := is.New(t)

		originalKind := &KindWithFieldWithNoClone{Field: "hey"}
		expectedKind := &KindWithFieldWithNoClone{Field: "hey"}
		err := erk.New(originalKind, "my message")

		kindCopy, ok := err.(*erk.Error).Kind().(*KindWithFieldWithNoClone)
		is.True(ok)
		is.Equal(kindCopy, expectedKind)

		kindCopy.Field = "something else"
		is.Equal(originalKind, expectedKind) // It should not modify the original kind
	})
}

func TestErrorWithParams(t *testing.T) {
	t.Run("with nil params, setting nil params", func(t *testing.T) {
		is := is.New(t)

		err1 := erk.New(ErkExample{}, "my message")
		err2 := err1.(*erk.Error).WithParams(nil)
		is.Equal(err2, err1)
		is.Equal(err2.(*erk.Error).Params(), nil)
	})

	t.Run("with nil params, setting two params", func(t *testing.T) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message")
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"a": "hello", "b": "world"})
	})

	t.Run("with present params, setting nil params", func(t *testing.T) {
		is := is.New(t)

		err1 := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err2 := err1.(*erk.Error).WithParams(nil)
		is.Equal(err2, err1)
		is.Equal(err2.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("with present params, setting two params", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
	})

	t.Run("with present params, deleting one param", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = err.(*erk.Error).WithParams(erk.Params{"a": "hello", "b": "world", "1": nil})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "a": "hello", "b": "world"})
	})

	t.Run("params are cloned", func(t *testing.T) {
		is := is.New(t)

		originalErr := erk.NewWith(ErkExample{}, "my message", erk.Params{
			"param1": "param1 value",
		})

		modifiedErr := erk.WithParams(originalErr, erk.Params{
			"param2": "param2 value",
		})

		is.Equal(erk.GetParams(originalErr), erk.Params{
			"param1": "param1 value",
		}) // The original error params should not be modified

		is.Equal(erk.GetParams(modifiedErr), erk.Params{
			"param1": "param1 value",
			"param2": "param2 value",
		})
	})
}

func TestErrorParams(t *testing.T) {
	t.Run("returns parameters", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("returns a copy of the parameters", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		params := err.(*erk.Error).Params()
		params["0"] = "changed"
		is.Equal(err.(*erk.Error).Params(), erk.Params{"0": "hey", "1": "there"})
	})
}

func TestErrorExport(t *testing.T) {
	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)
		is.Equal(errc.Kind, "github.com/JosiahWitt/erk_test:ErkExample")
		is.Equal(errc.Message, "my message: the world")
		is.Equal(errc.Params, erk.Params{"a": "the world"})
	})

	t.Run("returns a copy", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)
		errc.Params["a"] = "123"
		is.Equal(erk.GetParams(err), erk.Params{"a": "the world"})
	})

	t.Run("to JSON", func(t *testing.T) {
		t.Run("with valid params", func(t *testing.T) {
			is := is.New(t)

			val := "the world"
			err := erk.New(ErkExample{}, "my message: {{.a}}")
			err = erk.WithParam(err, "a", val)
			errc := err.(*erk.Error).Export().(*erk.ExportedError)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"}}`)
		})

		t.Run("with no params", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			errc := err.(*erk.Error).Export().(*erk.ExportedError)
			b, jerr := json.Marshal(errc)
			is.NoErr(jerr)
			is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message"}`)
		})
	})
}

func TestErrorMarshalJSON(t *testing.T) {
	t.Run("with valid params", func(t *testing.T) {
		is := is.New(t)

		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		b, jerr := json.Marshal(err)
		is.NoErr(jerr)
		is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"}}`)
	})
}

func withStrictMode(enabled bool, fn func()) {
	erkstrict.SetStrictMode(enabled)
	defer erkstrict.SetStrictMode(false)
	fn()
}

type KindWithFieldWithNoClone struct {
	Field string
}

func (k KindWithFieldWithNoClone) KindStringFor(erk.Kind) string {
	return k.Field
}
