package erkmock_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkmock"
	"github.com/matryer/is"
)

type TestKind struct {
	erk.DefaultKind
}

const (
	expectedKindString   = "github.com/JosiahWitt/erk/erkmock_test:TestKind"
	expectedErrorMessage = "MOCK: " + expectedKindString
)

func TestFor(t *testing.T) {
	is := is.New(t)

	m := erkmock.For(TestKind{})
	is.Equal(m.(erk.Kindable).Kind(), TestKind{})
	is.Equal(m.(erk.Paramable).Params(), erk.Params{})
}

func TestError(t *testing.T) {
	is := is.New(t)

	m := erkmock.For(TestKind{})
	is.Equal(m.Error(), expectedErrorMessage)
}

func TestExport(t *testing.T) {
	t.Run("with no params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})

		is.Equal(m.(erk.Exportable).Export(), &erk.ExportedError{
			BaseExport: erk.BaseExport{
				Kind:    expectedKindString,
				Message: expectedErrorMessage,
				Params:  erk.Params{},
			},
		})
	})

	t.Run("with params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		is.Equal(err, m)

		is.Equal(m.(erk.Exportable).Export(), &erk.ExportedError{
			BaseExport: erk.BaseExport{
				Kind:    expectedKindString,
				Message: expectedErrorMessage,
				Params: erk.Params{
					"param1": "hello",
				},
			},
		})
	})
}

func TestIs(t *testing.T) {
	type AnotherTestKind struct {
		erk.DefaultKind
	}

	t.Run("identity", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		is.True(errors.Is(m, m))
	})

	t.Run("two mocks with the same kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(TestKind{})
		is.True(errors.Is(m1, m2))
		is.True(errors.Is(m2, m1))
	})

	t.Run("two mocks with different kinds", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(AnotherTestKind{})
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})

	t.Run("erk error with same kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erk.New(TestKind{}, "my message")
		is.True(errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1)) // From the erk error's perspective the mock is not equivalent
	})

	t.Run("erk error with different kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erk.New(AnotherTestKind{}, "my message")
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})
}

func TestWithParams(t *testing.T) {
	t.Run("setting params once", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		is.Equal(err, m)

		is.Equal(m.(erk.Paramable).Params(), erk.Params{
			"param1": "hello",
		})
	})

	t.Run("setting params more than once", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		is.Equal(err, m)

		err2 := m.(erk.Paramable).WithParams(erk.Params{
			"param2": "hello 2",
			"param3": "hello 3",
		})
		is.Equal(err2, m)

		is.Equal(m.(erk.Paramable).Params(), erk.Params{
			"param1": "hello",
			"param2": "hello 2",
			"param3": "hello 3",
		})
	})

	t.Run("overwriting params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
			"param2": "hello 2",
		})
		is.Equal(err, m)

		err2 := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello - updated",
			"param3": "hello 3",
		})
		is.Equal(err2, m)

		is.Equal(m.(erk.Paramable).Params(), erk.Params{
			"param1": "hello - updated",
			"param2": "hello 2",
			"param3": "hello 3",
		})
	})
}
