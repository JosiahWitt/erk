package erkjson_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/JosiahWitt/erk/erkjson"
	"github.com/matryer/is"
)

func TestSetJSONError(t *testing.T) {
	is := is.New(t)
	const jsonError = `{"this": "is", "json": "!"}`

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetJSONError(jsonError)
	is.Equal(wrapper.Error(), jsonError)
}

func TestSetOriginalError(t *testing.T) {
	is := is.New(t)
	originalError := errors.New("something bad happened")

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetOriginalError(originalError)
	is.Equal(wrapper.Unwrap(), originalError)
}

func TestUnwrap(t *testing.T) {
	is := is.New(t)
	originalError := errors.New("something bad happened")

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetOriginalError(originalError)
	is.Equal(errors.Unwrap(wrapper), originalError)
}

func TestIsNil(t *testing.T) {
	t.Run("with nil JSONWrapper", func(t *testing.T) {
		is := is.New(t)

		var wrapper *erkjson.JSONWrapper
		is.True(wrapper.IsNil())
	})

	t.Run("with non nil JSONWrapper", func(t *testing.T) {
		is := is.New(t)

		wrapper := &erkjson.JSONWrapper{}
		is.Equal(wrapper.IsNil(), false)
	})
}

func TestMarshalJSON(t *testing.T) {
	is := is.New(t)
	const jsonError = `{"this":"is","json":"!"}`

	wrapper := &erkjson.JSONWrapper{}
	wrapper.SetJSONError(jsonError)

	marshalJSON, err := json.Marshal(wrapper)
	is.NoErr(err)
	is.Equal(string(marshalJSON), jsonError)
}

func TestExportError(t *testing.T) {
	type Entry struct {
		Name                string
		Error               error
		ExpectedErrorString string
		TypeCheck           func(is *is.I, entry *Entry, exportedError error)
	}

	table := []Entry{
		{
			Name:                "with pointer",
			Error:               erk.New(&TestPtrWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_ptr_wrapable_kind","message":"my message"}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				is.True(ok)
			},
		},
		{
			Name:                "with pointer but not wrapable",
			Error:               erk.New(&TestPtrNonWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_ptr_non_wrapable_kind","message":"my message"}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper)
				is.True(ok)
			},
		},
		{
			Name:                "with value",
			Error:               erk.New(TestValueWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_value_wrapable_kind","message":"my message"}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since it's a value, it doesn't satisfy the interface
				is.True(ok)
			},
		},
		{
			Name:                "with value embedding nil pointer",
			Error:               erk.New(TestValueWithPtrWrapableKind{}, "my message"),
			ExpectedErrorString: `{"kind":"test_value_with_ptr_wrapable_kind","message":"my message"}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since the pointer is nil, it returns the JSONWrapper type
				is.True(ok)
			},
		},
		{
			Name:                "with nil kind",
			Error:               erk.New(nil, "my message"),
			ExpectedErrorString: `{"kind":null,"message":"my message"}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*erkjson.JSONWrapper) // Since the kind is nil, it returns the JSONWrapper type
				is.True(ok)
			},
		},
		{
			Name:  "with unmarshalable error",
			Error: erk.NewWith(&TestPtrWrapableKind{}, "my message", erk.Params{"invalid": make(chan struct{})}),
			ExpectedErrorString: `{"kind":"erk:error_is_invalid_json",` +
				`"message":"The error cannot be wrapped as JSON: json: error calling MarshalJSON for type *erk.Error: ` +
				`json: error calling MarshalJSON for type erk.Params: json: unsupported type: chan struct {}"` +
				`,"params":{"err":"my message"}}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				is.True(ok)
			},
		},
		{
			Name: "with error group",
			Error: erg.New(&TestPtrWrapableKind{}, "my group",
				erk.New(&TestPtrWrapableKind{}, "my error"),
			),
			ExpectedErrorString: `{"kind":"test_ptr_wrapable_kind","message":"my group",` +
				`"errors":[{"kind":"test_ptr_wrapable_kind","message":"my error"}]}`,
			TypeCheck: func(is *is.I, entry *Entry, exportedError error) {
				_, ok := exportedError.(*TestPtrWrapableKind)
				is.True(ok)
			},
		},
	}

	for _, entry := range table {
		entry := entry // Pin range variable

		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			exportedError := erkjson.ExportError(entry.Error)
			is.Equal(exportedError.Error(), entry.ExpectedErrorString)
			entry.TypeCheck(is, &entry, exportedError)

			errWithParams := erk.WithParams(entry.Error, erk.Params{
				"param1": "value1",
			})
			erkjson.ExportError(errWithParams)                         // nolint:errcheck // Original could be modified if not properly cloned
			is.Equal(exportedError.Error(), entry.ExpectedErrorString) // Ensure original was not modified

			is.Equal(errors.Unwrap(exportedError), entry.Error) // Original error is wrapped
		})
	}
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
