package erkmock_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkmock"
	"github.com/matryer/is"
)

type TestKind struct {
	erk.DefaultKind
}

const (
	expectedKindString = "github.com/JosiahWitt/erk/erkmock_test:TestKind"
)

func TestFor(t *testing.T) {
	is := is.New(t)

	m := erkmock.For(TestKind{})
	is.Equal(m.(erk.Kindable).Kind(), TestKind{})
	is.Equal(m.(erk.Paramable).Params(), erk.Params{})
}

func TestSetMessage(t *testing.T) {
	is := is.New(t)

	m := erkmock.For(TestKind{})
	m.(*erkmock.Mock).SetMessage("my message")
	is.Equal(m.(erk.Exportable).ExportRawMessage(), "my message")
}

func TestError(t *testing.T) {
	t.Run("with basic mock error: with no params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		is.Equal(m.Error(), fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{}))
	})

	t.Run("with basic mock error: with params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		m = erk.WithParams(m, erk.Params{"param1": "abc", "param2": 123})
		is.Equal(m.Error(), fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "abc", "param2": 123}))
	})

	t.Run("with mock error with message: with no params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.From(erk.New(TestKind{}, "my message"))
		is.Equal(m.Error(), fmt.Sprintf("{KIND: \"%s\", RAW MESSAGE: \"my message\", PARAMS: %+v}", expectedKindString, erk.Params{}))
	})

	t.Run("with mock error with message: with params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.From(erk.New(TestKind{}, "my message"))
		m = erk.WithParams(m, erk.Params{"param1": "abc", "param2": 123})
		is.Equal(m.Error(), fmt.Sprintf("{KIND: \"%s\", RAW MESSAGE: \"my message\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "abc", "param2": 123}))
	})
}

func TestExportRawMessage(t *testing.T) {
	is := is.New(t)

	m := erkmock.For(TestKind{})
	m.(*erkmock.Mock).SetMessage("my message")
	is.Equal(m.(erk.Exportable).ExportRawMessage(), "my message")
}

func TestExport(t *testing.T) {
	t.Run("with no params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})

		is.Equal(m.(erk.Exportable).Export(), &erk.BaseExport{
			Kind:    expectedKindString,
			Message: fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{}),
			Params:  erk.Params{},
		})
	})

	t.Run("with nil kind", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(nil)

		is.Equal(m.(erk.Exportable).Export(), &erk.BaseExport{
			Kind:    "",
			Message: fmt.Sprintf("{KIND: \"\", PARAMS: %+v}", erk.Params{}),
			Params:  erk.Params{},
		})
	})

	t.Run("with params", func(t *testing.T) {
		is := is.New(t)

		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		is.Equal(err, m)

		is.Equal(m.(erk.Exportable).Export(), &erk.BaseExport{
			Kind:    expectedKindString,
			Message: fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "hello"}),
			Params: erk.Params{
				"param1": "hello",
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

	t.Run("no message: two mocks with the same kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(TestKind{})
		is.True(errors.Is(m1, m2))
		is.True(errors.Is(m2, m1))
	})

	t.Run("no message: two mocks with different kinds", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(AnotherTestKind{})
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})

	t.Run("no message: erk error with same kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erk.New(TestKind{}, "my message")
		is.True(errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1)) // From the erk error's perspective the mock is not equivalent
	})

	t.Run("no message: erk error with different kind", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.For(TestKind{})
		m2 := erk.New(AnotherTestKind{}, "my message")
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})

	t.Run("with message: erk error with same kind different message", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.From(erk.New(TestKind{}, "my message 1"))
		m2 := erk.New(TestKind{}, "my message 2")
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})

	t.Run("with message: erk error with different kind same message", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.From(erk.New(TestKind{}, "my message"))
		m2 := erk.New(AnotherTestKind{}, "my message")
		is.True(!errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1))
	})

	t.Run("with message: erk error with same kind same message", func(t *testing.T) {
		is := is.New(t)

		m1 := erkmock.From(erk.New(TestKind{}, "my message"))
		m2 := erk.New(TestKind{}, "my message")
		is.True(errors.Is(m1, m2))
		is.True(!errors.Is(m2, m1)) // From the erk error's perspective the mock is not equivalent
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
