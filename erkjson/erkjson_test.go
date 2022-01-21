package erkjson_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/JosiahWitt/erk/erkjson"
)

func TestSetJSONError(t *testing.T) {
	ensure := ensure.New(t)
	const jsonError = `{"this": "is", "json": "!"}`

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetJSONError(jsonError)
	ensure(wrapper.Error()).Equals(jsonError)
}

func TestSetOriginalError(t *testing.T) {
	ensure := ensure.New(t)
	originalError := errors.New("something bad happened")

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetOriginalError(originalError)
	ensure(wrapper.Unwrap()).Equals(originalError)
}

func TestUnwrap(t *testing.T) {
	ensure := ensure.New(t)
	originalError := errors.New("something bad happened")

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetOriginalError(originalError)
	ensure(errors.Unwrap(wrapper)).Equals(originalError)
}

func TestIsNil(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with nil JSONWrapper", func(ensure ensurepkg.Ensure) {
		var wrapper *erkjson.JSONWrapper
		ensure(wrapper.IsNil()).IsTrue()
	})

	ensure.Run("with non nil JSONWrapper", func(ensure ensurepkg.Ensure) {
		wrapper := &erkjson.JSONWrapper{}
		ensure(wrapper.IsNil()).IsFalse()
	})
}

func TestMarshalJSON(t *testing.T) {
	ensure := ensure.New(t)
	const jsonError = `{"this":"is","json":"!"}`

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetJSONError(jsonError)

	marshalJSON, err := json.Marshal(wrapper)
	ensure(err).IsNotError()
	ensure(string(marshalJSON)).Equals(jsonError)
}

func TestExportError(t *testing.T) {
	ensure := ensure.New(t)

	type Entry struct {
		Name                string
		Error               error
		ExpectedErrorString string
		TypeCheck           func(ensure ensurepkg.Ensure, entry *Entry, exportedError error)
	}

	table := []Entry{
		{
			Name:                "with pointer",
			Error:               erk.New(&TestPtrWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_ptr_wrapable_kind","message":"my message"}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				ensure(ok).IsTrue()
			},
		},
		{
			Name:                "with pointer but not wrapable",
			Error:               erk.New(&TestPtrNonWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_ptr_non_wrapable_kind","message":"my message"}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper)
				ensure(ok).IsTrue()
			},
		},
		{
			Name:                "with value",
			Error:               erk.New(TestValueWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_value_wrapable_kind","message":"my message"}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since it's a value, it doesn't satisfy the interface
				ensure(ok).IsTrue()
			},
		},
		{
			Name:                "with value embedding nil pointer",
			Error:               erk.New(TestValueWithPtrWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_value_with_ptr_wrapable_kind","message":"my message"}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since the pointer is nil, it returns the JSONWrapper type
				ensure(ok).IsTrue()
			},
		},
		{
			Name:                "with nil kind",
			Error:               erk.New(nil, "my message"),
			ExpectedErrorString: `{"kind":null,"message":"my message"}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since the kind is nil, it returns the JSONWrapper type
				ensure(ok).IsTrue()
			},
		},
		{
			Name:  "with unmarshalable error",
			Error: erk.NewWith(&TestPtrWrapableKind{}, "my message", erk.Params{"invalid": make(chan struct{})}),
			ExpectedErrorString: `{"kind":"erk:error_is_invalid_json",` +
				`"message":"The error cannot be wrapped as JSON: json: error calling MarshalJSON for type *erk.Error: ` +
				`json: error calling MarshalJSON for type erk.Params: json: unsupported type: chan struct {}"` +
				`,"params":{"err":"my message"}}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				ensure(ok).IsTrue()
			},
		},
		{
			Name: "with error group",
			Error: erg.New(&TestPtrWrapableKind{}, "my group",
				erk.New(&TestPtrWrapableKind{}, "my error"),
			),
			ExpectedErrorString: `{"kind":"test_ptr_wrapable_kind","message":"my group",` +
				`"errors":[{"kind":"test_ptr_wrapable_kind","message":"my error"}]}`,
			TypeCheck: func(ensure ensurepkg.Ensure, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				ensure(ok).IsTrue()
			},
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]

		exportedError := erkjson.ExportError(entry.Error)
		ensure(exportedError.Error()).Equals(entry.ExpectedErrorString)
		entry.TypeCheck(ensure, &entry, exportedError)

		errWithParams := erk.WithParams(entry.Error, erk.Params{
			"param1": "value1",
		})
		erkjson.ExportError(errWithParams)                              // nolint:errcheck // Original could be modified if not properly cloned
		ensure(exportedError.Error()).Equals(entry.ExpectedErrorString) // Ensure original was not modified

		ensure(errors.Unwrap(exportedError)).IsError(entry.Error) // Original error is wrapped
	})
}

type TestValueWithPtrWrapableKind struct {
	erk.DefaultKind
	*erkjson.JSONWrapper
}

func (TestValueWithPtrWrapableKind) KindStringFor(kind erk.Kind) string {
	return "test_value_with_ptr_wrapable_kind"
}

type TestValueWrapableKind struct {
	erk.DefaultKind
	erkjson.JSONWrapper
}

func (TestValueWrapableKind) KindStringFor(kind erk.Kind) string {
	return "test_value_wrapable_kind"
}

type TestPtrWrapableKind struct {
	erk.DefaultPtrKind
	erkjson.JSONWrapper
}

func (*TestPtrWrapableKind) KindStringFor(kind erk.Kind) string {
	return "test_ptr_wrapable_kind"
}

type TestPtrNonWrapableKind struct {
	erk.DefaultPtrKind
}

func (*TestPtrNonWrapableKind) KindStringFor(kind erk.Kind) string {
	return "test_ptr_non_wrapable_kind"
}
