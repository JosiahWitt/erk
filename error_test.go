package erk_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
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
	t.Helper()

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

						isValid := regexp.MustCompile(templateInvalidRegexp).MatchString(str)
						is.True(isValid)
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

					isValid := regexp.MustCompile(templateInvalidRegexp).MatchString(str)
					is.True(isValid)
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

					isValid := regexp.MustCompile(templateInvalidParamErrorRegexp).MatchString(str)
					is.True(isValid)
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

					isValid := regexp.MustCompile(templateMissingParamErrorRegexp).MatchString(str)
					is.True(isValid)
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

func TestExportRawMessage(t *testing.T) {
	is := is.New(t)

	err := erk.New(ErkExample{}, "my message {{.key}}")
	is.Equal(err.(*erk.Error).ExportRawMessage(), "my message {{.key}}")
}

func TestErrorExport(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with valid params", func(ensure ensurepkg.Ensure) {
		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(strPtr("github.com/JosiahWitt/erk_test:ErkExample"))
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})
		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{})
	})

	ensure.Run("returns a copy", func(ensure ensurepkg.Ensure) {
		val := "the world"
		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)
		errc.Params["a"] = "123"
		ensure(erk.GetParams(err)).Equals(erk.Params{"a": "the world"})
	})

	ensure.Run("with a non-erk error", func(ensure ensurepkg.Ensure) {
		originalErr := errors.New("original error")
		errc := erk.Export(originalErr).(*erk.ExportedError)

		ensure(errc.Kind).Equals(nil)
		ensure(errc.Type).Equals(strPtr("errors:errorString"))
		ensure(errc.Message).Equals("original error")
		ensure(errc.Params).Equals(erk.Params{})
		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{})
	})

	ensure.Run("with a nil kind string", func(ensure ensurepkg.Ensure) {
		val := "the world"
		err := erk.New(nil, "my message: {{.a}}")
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(nil)
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})
		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{})
	})

	ensure.Run("with a wrapped error", func(ensure ensurepkg.Ensure) {
		val := "the world"
		originalErr := errors.New("original error")
		err := erk.Wrap(ErkExample{}, "my message: {{.a}}", originalErr)
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(strPtr("github.com/JosiahWitt/erk_test:ErkExample"))
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})
		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{
			&erk.ExportedError{
				Kind:    nil,
				Type:    strPtr("errors:errorString"),
				Message: "original error",
			},
		})
	})

	ensure.Run("with a wrapped erkable", func(ensure ensurepkg.Ensure) {
		val := "the world"
		originalErr := &SimpleErkable{}
		err := erk.Wrap(ErkExample{}, "my message: {{.a}}", originalErr)
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(strPtr("github.com/JosiahWitt/erk_test:ErkExample"))
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})
		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{(&SimpleErkable{}).Export()})
	})

	ensure.Run("with a doubly wrapped erk error", func(ensure ensurepkg.Ensure) {
		val := "the world"
		originalErr := errors.New("original error")
		midErr := erk.Wrap(ErkExample2{}, "in the middle", originalErr)
		err := erk.Wrap(ErkExample{}, "my message: {{.a}}", midErr)
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(strPtr("github.com/JosiahWitt/erk_test:ErkExample"))
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})

		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{
			&erk.ExportedError{
				Kind:    strPtr("github.com/JosiahWitt/erk_test:ErkExample2"),
				Message: "in the middle",
			},
			&erk.ExportedError{
				Kind:    nil,
				Type:    strPtr("errors:errorString"),
				Message: "original error",
			},
		})
	})

	ensure.Run("with a doubly wrapped non-erk error", func(ensure ensurepkg.Ensure) {
		val := "the world"
		originalErr := errors.New("original error")
		midErr := fmt.Errorf("in the middle: %w", originalErr)
		err := erk.Wrap(ErkExample{}, "my message: {{.a}}", midErr)
		err = erk.WithParam(err, "a", val)
		errc := err.(*erk.Error).Export().(*erk.ExportedError)

		ensure(errc.Kind).Equals(strPtr("github.com/JosiahWitt/erk_test:ErkExample"))
		ensure(errc.Type).IsNil()
		ensure(errc.Message).Equals("my message: the world")
		ensure(errc.Params).Equals(erk.Params{"a": "the world"})

		ensure(errc.ErrorStack).Equals([]erk.ExportedErkable{
			&erk.ExportedError{
				Kind:    nil,
				Type:    strPtr("fmt:wrapError"),
				Message: "in the middle: original error",
			},
			&erk.ExportedError{
				Kind:    nil,
				Type:    strPtr("errors:errorString"),
				Message: "original error",
			},
		})
	})
}

func TestErrorMarshalJSON(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with valid params", func(ensure ensurepkg.Ensure) {
		is := is.New(t)

		err := erk.New(ErkExample{}, "my message: {{.a}}")
		err = erk.WithParam(err, "a", "the world")
		b, jerr := json.Marshal(err)
		is.NoErr(jerr)
		is.Equal(string(b), `{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"}}`)
	})

	ensure.Run("with no params", func(ensure ensurepkg.Ensure) {
		err := erk.New(ErkExample{}, "my message")
		b, jerr := json.Marshal(err)
		ensure(jerr).IsNotError()
		ensure(string(b)).Equals(`{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message"}`)
	})

	ensure.Run("with doubly wrapped error and params", func(ensure ensurepkg.Ensure) {
		originalErr := errors.New("original error")
		midErr := erk.Wrap(ErkExample2{}, "in the middle", originalErr)
		midErr = erk.WithParam(midErr, "stuck", true)
		err := erk.Wrap(ErkExample{}, "my message: {{.a}}", midErr)
		err = erk.WithParam(err, "a", "the world")

		b, jerr := json.Marshal(err)
		ensure(jerr).IsNotError()
		ensure(string(b)).Equals(
			`{"kind":"github.com/JosiahWitt/erk_test:ErkExample","message":"my message: the world","params":{"a":"the world"},` +
				`"errorStack":[{"kind":"github.com/JosiahWitt/erk_test:ErkExample2","message":"in the middle","params":{"stuck":true}},` +
				`{"kind":null,"type":"errors:errorString","message":"original error"}]}`,
		)
	})

	ensure.Run("with a non-erk error", func(ensure ensurepkg.Ensure) {
		originalErr := errors.New("original error")
		b, jerr := json.Marshal(erk.Export(originalErr))
		ensure(jerr).IsNotError()
		ensure(string(b)).Equals(`{"kind":null,"type":"errors:errorString","message":"original error"}`)
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

type SimpleErkable struct{}

var _ erk.Erkable = &SimpleErkable{}

func (e *SimpleErkable) Error() string               { return "simple erkable" }
func (e *SimpleErkable) ExportRawMessage() string    { return e.Error() }
func (e *SimpleErkable) Kind() erk.Kind              { return ErkExample{} }
func (e *SimpleErkable) Params() erk.Params          { return erk.Params{} }
func (e *SimpleErkable) WithParams(erk.Params) error { return e }
func (e *SimpleErkable) Export() erk.ExportedErkable {
	return &erk.BaseExport{Message: "exported simple erkable"}
}

func strPtr(str string) *string {
	return &str
}

const (
	disclosureRegexp = "NOTE: This message was raised because strict mode is enabled. " +
		"Strict mode is automatically enabled in tests. " +
		"To disable strict mode in tests, set the environment variable ERK_STRICT_MODE=false or use `erkstrict.SetStrictMode\\(false\\)`. " +
		"It is recommended to use strict mode for testing and development, to catch when an error message is invalid. " +
		"If you are attempting to return an error from a mock, you can use `erkmock.From\\(err\\)` to bypass strict mode.\\n\\n" +
		"\\*{25}\\n"

	templateInvalidParamErrorRegexp = "\\n\\*{25}\\n\\n" +
		"Unable to execute error template:\\n" +
		"\\tKind: github.com/JosiahWitt/erk_test:ErkExample\\n" +
		"\\tTemplate: my message {{call .a}}\\n" +
		"\\tParams: map\\[a:.+\\]\\n" +
		"\\tError:.+error calling call.+\\n\\n" +
		disclosureRegexp

	templateMissingParamErrorRegexp = "\\n\\*{25}\\n\\n" +
		"Unable to execute error template:\\n" +
		"\\tKind: github.com/JosiahWitt/erk_test:ErkExample\\n" +
		"\\tTemplate: my message: {{.a}}, {{.b}}!\\n" +
		"\\tParams: map\\[a:hello]\\n" +
		"\\tError:.+map has no entry for key \"b\"\\n\\n" +
		disclosureRegexp

	templateInvalidRegexp = "\\n\\*{25}\\n\\n" +
		"Unable to parse error template:\\n" +
		"\\tKind: github.com/JosiahWitt/erk_test:ErkExample\\n" +
		"\\tTemplate: my message {{}}}\\n" +
		"\\tError:.+missing value for command\\n\\n" +
		disclosureRegexp
)
