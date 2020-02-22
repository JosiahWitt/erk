package erk_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type (
	ErkExample  struct{ erk.DefaultKind }
	ErkExample2 struct{ erk.DefaultKind }
)

func TestNew(t *testing.T) {
	is := is.New(t)

	msg := "my message"
	err := erk.New(ErkExample{}, msg)
	is.Equal(err.Error(), msg)
	is.Equal(erk.GetParams(err), nil)
	is.Equal(erk.GetKind(err), ErkExample{})
}

func TestNewWith(t *testing.T) {
	is := is.New(t)

	msg := "my message: {{.a}}, {{.b}}!"
	err := erk.NewWith(ErkExample{}, msg, erk.Params{"a": "hello", "b": "world"})
	is.Equal(err.Error(), "my message: hello, world!")
	is.Equal(erk.GetParams(err), erk.Params{"a": "hello", "b": "world"})
	is.Equal(erk.GetKind(err), ErkExample{})
}

func TestError(t *testing.T) {
	t.Run("with invalid template", func(t *testing.T) {
		withErkStrictEnv("false", func() {
			is := is.New(t)

			msg := "my message {{}}}"
			err := erk.New(ErkExample{}, msg)
			is.Equal(err.Error(), msg)
		})
	})

	t.Run("with invalid param", func(t *testing.T) {
		withErkStrictEnv("false", func() {
			is := is.New(t)

			msg := "my message {{call .a}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", func() { panic("just testing") })
			is.Equal(err.Error(), msg)
		})
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
		withErkStrictEnv("false", func() {
			is := is.New(t)

			msg := "my message: {{.a}}, {{.b}}!"
			err := erk.New(ErkExample{}, msg)
			err = erk.WithParam(err, "a", "hello")
			is.Equal(err.Error(), "my message: hello, <no value>!")
		})
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

			wrappedErr := errors.New("see! there are no newlines; this one (\\n) is escaped!")
			msg := "my message: {{.err}}"
			err := erk.New(ErkExample{}, msg)
			err = erk.WrapAs(err, wrappedErr)
			is.Equal(err.Error(), "my message: see! there are no newlines; this one (\\n) is escaped!")
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
			withErkStrictEnv("false", func() {
				is := is.New(t)

				msg := "my message: {{.err}}"
				err := erk.New(ErkExample{}, msg)
				is.Equal(err.Error(), "my message: <no value>")
			})
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
	expectNoPanic := func(t *testing.T) func() {
		return func() {
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
		}
	}

	expectPanic := func(t *testing.T) func() {
		return func() {
			t.Run("with invalid template", func(t *testing.T) {
				is := is.New(t)

				defer func() {
					if res := recover(); res != nil {
						str, ok := res.(string)
						is.True(ok)

						is.True(strings.Contains(str, "Unable to parse error template"))
						is.True(strings.Contains(str, "Template: my message {{}}}"))
						is.True(strings.Contains(str, "Error: "))
						is.Equal(os.Getenv("ERK_STRICT"), "true")
					}
				}()

				msg := "my message {{}}}"
				err := erk.New(ErkExample{}, msg)
				err.Error()
				is.Fail() // Expected panic
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
						is.Equal(os.Getenv("ERK_STRICT"), "true")
					}
				}()

				msg := "my message {{call .a}}"
				err := erk.New(ErkExample{}, msg)
				err = erk.WithParam(err, "a", func() { panic("just testing") })
				err.Error()
				is.Fail() // Expected panic
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
						is.Equal(os.Getenv("ERK_STRICT"), "true")
					}
				}()

				msg := "my message: {{.a}}, {{.b}}!"
				err := erk.New(ErkExample{}, msg)
				err = erk.WithParam(err, "a", "hello")
				err.Error()
				is.Fail() // Expected panic
			})
		}
	}

	t.Run("not in tests", func(t *testing.T) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing"}
		defer func() {
			os.Args = originalArgs
		}()

		t.Run("with no ERK_STRICT environment variable", func(t *testing.T) {
			withErkStrictEnv("", expectNoPanic(t))
		})

		t.Run("with ERK_STRICT=false", func(t *testing.T) {
			withErkStrictEnv("false", expectNoPanic(t))
		})

		t.Run("with ERK_STRICT=true", func(t *testing.T) {
			withErkStrictEnv("true", expectPanic(t))
		})
	})

	t.Run("in tests", func(t *testing.T) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing", "-test.thing=true"}
		defer func() {
			os.Args = originalArgs
		}()

		t.Run("with no ERK_STRICT environment variable", func(t *testing.T) {
			withErkStrictEnv("", expectPanic(t))
		})

		t.Run("with ERK_STRICT=false", func(t *testing.T) {
			withErkStrictEnv("false", expectNoPanic(t))
		})

		t.Run("with ERK_STRICT=true", func(t *testing.T) {
			withErkStrictEnv("true", expectPanic(t))
		})
	})

	t.Run("when not changing args (in test)", func(t *testing.T) {
		t.Run("with no ERK_STRICT environment variable", func(t *testing.T) {
			withErkStrictEnv("", expectPanic(t))
		})

		t.Run("with ERK_STRICT=false", func(t *testing.T) {
			withErkStrictEnv("false", expectNoPanic(t))
		})

		t.Run("with ERK_STRICT=true", func(t *testing.T) {
			withErkStrictEnv("true", expectPanic(t))
		})
	})
}

func TestIs(t *testing.T) {
	table := []struct {
		Name   string
		Error1 error
		Error2 error
		Equal  bool
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
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)
			is.Equal(errors.Is(entry.Error1, entry.Error2), entry.Equal)
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
	is := is.New(t)

	err := erk.New(ErkExample{}, "my message")
	is.Equal(err.(*erk.Error).Kind(), ErkExample{})
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

func withErkStrictEnv(value string, fn func()) {
	strict, isSet := os.LookupEnv("ERK_STRICT")
	if value != "" {
		os.Setenv("ERK_STRICT", value)
	} else {
		os.Unsetenv("ERK_STRICT")
	}

	fn()

	if isSet {
		os.Setenv("ERK_STRICT", strict)
	} else {
		os.Unsetenv("ERK_STRICT")
	}
}
