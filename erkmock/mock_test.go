package erkmock_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkmock"
)

type TestKind struct {
	erk.DefaultKind
}

const (
	expectedKindString = "github.com/JosiahWitt/erk/erkmock_test:TestKind"
)

func TestFor(t *testing.T) {
	ensure := ensure.New(t)

	m := erkmock.For(TestKind{})
	ensure(m.(erk.Kindable).Kind()).Equals(TestKind{})
	ensure(m.(erk.Paramable).Params()).Equals(erk.Params{})
}

func TestSetMessage(t *testing.T) {
	ensure := ensure.New(t)

	m := erkmock.For(TestKind{})
	m.(*erkmock.Mock).SetMessage("my message")
	ensure(m.(erk.Exportable).ExportRawMessage()).Equals("my message")
}

func TestError(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with basic mock error: with no params", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		ensure(m.Error()).Equals(fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{}))
	})

	ensure.Run("with basic mock error: with params", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		m = erk.WithParams(m, erk.Params{"param1": "abc", "param2": 123})
		ensure(m.Error()).Equals(fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "abc", "param2": 123}))
	})

	ensure.Run("with mock error with message: with no params", func(ensure ensurepkg.Ensure) {
		m := erkmock.From(erk.New(TestKind{}, "my message"))
		ensure(m.Error()).Equals(fmt.Sprintf("{KIND: \"%s\", RAW MESSAGE: \"my message\", PARAMS: %+v}", expectedKindString, erk.Params{}))
	})

	ensure.Run("with mock error with message: with params", func(ensure ensurepkg.Ensure) {
		m := erkmock.From(erk.New(TestKind{}, "my message"))
		m = erk.WithParams(m, erk.Params{"param1": "abc", "param2": 123})
		ensure(m.Error()).Equals(fmt.Sprintf("{KIND: \"%s\", RAW MESSAGE: \"my message\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "abc", "param2": 123}))
	})
}

func TestExportRawMessage(t *testing.T) {
	ensure := ensure.New(t)

	m := erkmock.For(TestKind{})
	m.(*erkmock.Mock).SetMessage("my message")
	ensure(m.(erk.Exportable).ExportRawMessage()).Equals("my message")
}

func TestExport(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with no params", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})

		ensure(m.(erk.Exportable).Export()).Equals(&erk.BaseExport{
			Kind:    expectedKindString,
			Message: fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{}),
			Params:  erk.Params{},
		})
	})

	ensure.Run("with nil kind", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(nil)

		ensure(m.(erk.Exportable).Export()).Equals(&erk.BaseExport{
			Kind:    "",
			Message: fmt.Sprintf("{KIND: \"\", PARAMS: %+v}", erk.Params{}),
			Params:  erk.Params{},
		})
	})

	ensure.Run("with params", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		ensure(err).Equals(m)

		ensure(m.(erk.Exportable).Export()).Equals(&erk.BaseExport{
			Kind:    expectedKindString,
			Message: fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", expectedKindString, erk.Params{"param1": "hello"}),
			Params: erk.Params{
				"param1": "hello",
			},
		})
	})
}

func TestIs(t *testing.T) {
	ensure := ensure.New(t)

	type AnotherTestKind struct {
		erk.DefaultKind
	}

	ensure.Run("identity", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		ensure(errors.Is(m, m)).IsTrue()
	})

	ensure.Run("no message: two mocks with the same kind", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(TestKind{})
		ensure(errors.Is(m1, m2)).IsTrue()
		ensure(errors.Is(m2, m1)).IsTrue()
	})

	ensure.Run("no message: two mocks with different kinds", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.For(TestKind{})
		m2 := erkmock.For(AnotherTestKind{})
		ensure(errors.Is(m1, m2)).IsFalse()
		ensure(errors.Is(m2, m1)).IsFalse()
	})

	ensure.Run("no message: erk error with same kind", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.For(TestKind{})
		m2 := erk.New(TestKind{}, "my message")
		ensure(errors.Is(m1, m2)).IsTrue()
		ensure(errors.Is(m2, m1)).IsFalse() // From the erk error's perspective the mock is not equivalent
	})

	ensure.Run("no message: erk error with different kind", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.For(TestKind{})
		m2 := erk.New(AnotherTestKind{}, "my message")
		ensure(errors.Is(m1, m2)).IsFalse()
		ensure(errors.Is(m2, m1)).IsFalse()
	})

	ensure.Run("with message: erk error with same kind different message", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.From(erk.New(TestKind{}, "my message 1"))
		m2 := erk.New(TestKind{}, "my message 2")
		ensure(errors.Is(m1, m2)).IsFalse()
		ensure(errors.Is(m2, m1)).IsFalse()
	})

	ensure.Run("with message: erk error with different kind same message", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.From(erk.New(TestKind{}, "my message"))
		m2 := erk.New(AnotherTestKind{}, "my message")
		ensure(errors.Is(m1, m2)).IsFalse()
		ensure(errors.Is(m2, m1)).IsFalse()
	})

	ensure.Run("with message: erk error with same kind same message", func(ensure ensurepkg.Ensure) {
		m1 := erkmock.From(erk.New(TestKind{}, "my message"))
		m2 := erk.New(TestKind{}, "my message")
		ensure(errors.Is(m1, m2)).IsTrue()
		ensure(errors.Is(m2, m1)).IsFalse() // From the erk error's perspective the mock is not equivalent
	})
}

func TestWithParams(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("setting params once", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		ensure(err).Equals(m)

		ensure(m.(erk.Paramable).Params()).Equals(erk.Params{
			"param1": "hello",
		})
	})

	ensure.Run("setting params more than once", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
		})
		ensure(err).Equals(m)

		err2 := m.(erk.Paramable).WithParams(erk.Params{
			"param2": "hello 2",
			"param3": "hello 3",
		})
		ensure(err2).Equals(m)

		ensure(m.(erk.Paramable).Params()).Equals(erk.Params{
			"param1": "hello",
			"param2": "hello 2",
			"param3": "hello 3",
		})
	})

	ensure.Run("overwriting params", func(ensure ensurepkg.Ensure) {
		m := erkmock.For(TestKind{})
		err := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello",
			"param2": "hello 2",
		})
		ensure(err).Equals(m)

		err2 := m.(erk.Paramable).WithParams(erk.Params{
			"param1": "hello - updated",
			"param3": "hello 3",
		})
		ensure(err2).Equals(m)

		ensure(m.(erk.Paramable).Params()).Equals(erk.Params{
			"param1": "hello - updated",
			"param2": "hello 2",
			"param3": "hello 3",
		})
	})
}
