package erg_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/matryer/is"
)

const MyKindString = "github.com/JosiahWitt/erk/erg_test:MyKind"

type MyKind struct{ erk.DefaultKind }

func TestNew(t *testing.T) {
	is := is.New(t)

	msg := "my message"
	errs := []error{errors.New("err1"), errors.New("err2")}
	err := erg.New(MyKind{}, msg, append(errs, nil)...)

	is.Equal(erk.GetKind(err), MyKind{})
	is.Equal(err.Error(), "my message:\n - err1\n - err2")
	is.Equal(erg.GetErrors(err), errs)
}

func TestNewAs(t *testing.T) {
	is := is.New(t)

	header := errors.New("my header")
	errs := []error{errors.New("err1"), errors.New("err2")}
	err := erg.NewAs(header, append(errs, nil)...)

	is.Equal(erk.GetKind(err), nil)
	is.Equal(err.Error(), "my header:\n - err1\n - err2")
	is.Equal(erg.GetErrors(err), errs)
}

func TestGroupHeader(t *testing.T) {
	is := is.New(t)

	header := errors.New("my header")
	err := erg.NewAs(header)

	errorGroup, ok := err.(erg.Groupable)
	is.True(ok)
	is.Equal(errorGroup.Header(), header)
}

func TestGroupError(t *testing.T) {
	t.Run("with no errs", func(t *testing.T) {
		t.Run("with trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message:"
			err := erg.New(MyKind{}, msg)
			is.Equal(err.Error(), msg)
		})

		t.Run("with no trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message"
			err := erg.New(MyKind{}, msg)
			is.Equal(err.Error(), msg)
		})
	})

	t.Run("with two errs", func(t *testing.T) {
		t.Run("with trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message:"
			errs := []error{errors.New("err1"), errors.New("err2")}
			err := erg.New(MyKind{}, msg, errs...)
			is.Equal(err.Error(), "my message:\n - err1\n - err2")
		})

		t.Run("with no trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message"
			errs := []error{errors.New("err1"), errors.New("err2")}
			err := erg.New(MyKind{}, msg, errs...)
			is.Equal(err.Error(), "my message:\n - err1\n - err2")
		})
	})

	t.Run("with message template", func(t *testing.T) {
		t.Run("with trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message {{.val}}:"
			errs := []error{errors.New("err1"), errors.New("err2")}
			err := erk.WithParam(erg.New(MyKind{}, msg, errs...), "val", "my-val")
			is.Equal(err.Error(), "my message my-val:\n - err1\n - err2")
		})

		t.Run("with no trailing :", func(t *testing.T) {
			is := is.New(t)

			msg := "my message {{.val}}"
			errs := []error{errors.New("err1"), errors.New("err2")}
			err := erk.WithParam(erg.New(MyKind{}, msg, errs...), "val", "my-val")
			is.Equal(err.Error(), "my message my-val:\n - err1\n - err2")
		})
	})

	t.Run("with nested group errors", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		ergNested2 := erg.New(MyKind{}, "deeply nested", errors.New("ergNested2 err1"), errors.New("ergNested2 err2"))
		ergNested1 := erg.New(MyKind{}, "nested", errors.New("ergNested1 err1"), ergNested2)
		errs := []error{errors.New("err1"), ergNested1, errors.New("err2")}
		err := erg.New(MyKind{}, msg, errs...)
		is.Equal(err.Error(),
			`my message:
 - err1
 - nested:
   - ergNested1 err1
   - deeply nested:
     - ergNested2 err1
     - ergNested2 err2
 - err2`,
		)
	})

	t.Run("with nested group errors and erk error", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		ergNested2 := erg.New(MyKind{}, "deeply nested", errors.New("ergNested2 err1"), errors.New("ergNested2 err2"))
		erkErrNested := erk.WrapAs(erk.New(MyKind{}, "my erk error: {{.err}}"), ergNested2)
		ergNested1 := erg.New(MyKind{}, "nested", errors.New("ergNested1 err1"), ergNested2, erkErrNested)
		erkErr := erk.WrapAs(erk.New(MyKind{}, "my erk error 2: {{.err}}"), ergNested1)
		errs := []error{errors.New("err1"), ergNested1, erkErr, errors.New("err2")}
		err := erg.New(MyKind{}, msg, errs...)
		is.Equal(err.Error(),
			`my message:
 - err1
 - nested:
   - ergNested1 err1
   - deeply nested:
     - ergNested2 err1
     - ergNested2 err2
   - my erk error: deeply nested:
     - ergNested2 err1
     - ergNested2 err2
 - my erk error 2: nested:
   - ergNested1 err1
   - deeply nested:
     - ergNested2 err1
     - ergNested2 err2
   - my erk error: deeply nested:
     - ergNested2 err1
     - ergNested2 err2
 - err2`,
		)
	})
}

func TestErrorsString(t *testing.T) {
	t.Run("with no indentation", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		errs := []error{errors.New("err1"), errors.New("err2")}
		err := erg.New(MyKind{}, msg, errs...)
		is.Equal(err.(erk.ErrorIndentable).IndentError(""), "my message:\n - err1\n - err2")
	})
}

func TestGroupIs(t *testing.T) {
	t.Run("with equal erk error", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		errs := []error{errors.New("err1"), errors.New("err2")}
		erkErr := erk.New(MyKind{}, msg)
		err := erg.NewAs(erkErr, errs...)

		is.True(errors.Is(err, erkErr))
	})

	t.Run("with not equal erk error", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		errs := []error{errors.New("err1"), errors.New("err2")}
		erkErr := erk.New(MyKind{}, msg)
		err := erg.NewAs(erkErr, errs...)

		erkErr2 := erk.New(MyKind{}, "msg two")
		is.Equal(errors.Is(err, erkErr2), false)
	})

	t.Run("with not equal other error", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		errs := []error{errors.New("err1"), errors.New("err2")}
		erkErr := erk.New(MyKind{}, msg)
		err := erg.NewAs(erkErr, errs...)

		err2 := errors.New("message two")
		is.Equal(errors.Is(err, err2), false)
	})

	t.Run("check against error inside group", func(t *testing.T) {
		t.Run("with errors.New() error", func(t *testing.T) {
			is := is.New(t)

			msg := "my message"
			err2 := errors.New("err2")
			errs := []error{errors.New("err1"), err2}
			erkErr := erk.New(MyKind{}, msg)
			err := erg.NewAs(erkErr, errs...)
			is.True(errors.Is(err, err2))
		})

		t.Run("with erk error", func(t *testing.T) {
			is := is.New(t)

			msg := "my message"
			err2 := erk.New(MyKind{}, "my err2 message")
			errs := []error{errors.New("err1"), err2}
			erkErr := erk.New(MyKind{}, msg)
			err := erg.NewAs(erkErr, errs...)
			is.True(errors.Is(err, err2))
		})

		t.Run("with error not in group", func(t *testing.T) {
			is := is.New(t)

			msg := "my message"
			errs := []error{errors.New("err1"), errors.New("err1")}
			erkErr := erk.New(MyKind{}, msg)
			err := erg.NewAs(erkErr, errs...)
			is.Equal(errors.Is(err, errors.New("err3")), false)
		})
	})
}

func TestGroupWithParams(t *testing.T) {
	is := is.New(t)

	msg := "my message"
	errs := []error{errors.New("err1"), errors.New("err2")}
	err := erg.New(MyKind{}, msg, errs...)
	err2 := erk.WithParam(err, "param1", "my param 1")
	err2 = erk.WithParams(err2, erk.Params{"param2": "my param 2"})

	is.Equal(erg.GetErrors(err2), errs) // Errors are not lost
	is.Equal(erk.GetParams(err2), erk.Params{"param1": "my param 1", "param2": "my param 2"})
	is.Equal(erk.GetParams(err), erk.Params{}) // Original group is not modified
}

func TestGroupKind(t *testing.T) {
	is := is.New(t)

	err := erg.New(MyKind{}, "my message")
	is.Equal(erk.GetKind(err), MyKind{})
}

func TestGroupExportRawMessage(t *testing.T) {
	t.Run("with erk header", func(t *testing.T) {
		is := is.New(t)

		err := erg.New(MyKind{}, "my message {{.key}}")
		is.Equal(err.(*erg.Group).ExportRawMessage(), "my message {{.key}}")
	})

	t.Run("with basic error header", func(t *testing.T) {
		is := is.New(t)

		err := erg.NewAs(errors.New("my message"))
		is.Equal(err.(*erg.Group).ExportRawMessage(), "my message")
	})
}

func TestGroupExport(t *testing.T) {
	ensure := ensure.New(t)

	header := erk.New(MyKind{}, "my message {{.val}}")

	simpleErr1 := errors.New("err1")
	simpleErr2 := errors.New("err2")

	erkErr1 := erk.NewWith(MyKind{}, "err1 {{.param}}", erk.Params{"param": "hello"})
	erkErr2 := erk.NewWith(MyKind{}, "err2 {{.param}}", erk.Params{"param": "world"})

	table := []struct {
		Name string

		Header       error
		NestedErrors []error

		ExpectedKind string

		ExpectedExportedErrors []erk.ExportedErkable
		ExpectedJSON           string
	}{
		{
			Name: "with simple non-erk errors",

			Header:       header,
			NestedErrors: []error{simpleErr1, simpleErr2},

			ExpectedKind: MyKindString,

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(simpleErr1),
				erk.Export(simpleErr2),
			},

			ExpectedJSON: `{"kind":"` + MyKindString + `",` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":null,"type":"errors:errorString","message":"err1"},` +
				`{"kind":null,"type":"errors:errorString","message":"err2"}]}`,
		},
		{
			Name: "with erk errors",

			Header:       header,
			NestedErrors: []error{erkErr1, erkErr2},

			ExpectedKind: MyKindString,

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(erkErr1),
				erk.Export(erkErr2),
			},

			ExpectedJSON: `{"kind":"` + MyKindString + `",` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":"` + MyKindString + `","message":"err1 hello","params":{"param":"hello"}},` +
				`{"kind":"` + MyKindString + `","message":"err2 world","params":{"param":"world"}}]}`,
		},
		{
			Name: "with mixed errors",

			Header:       header,
			NestedErrors: []error{simpleErr1, erk.WrapAs(erkErr2, erkErr1)},

			ExpectedKind: MyKindString,

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(simpleErr1),
				erk.Export(erk.WrapAs(erkErr2, erkErr1)),
			},

			ExpectedJSON: `{"kind":"` + MyKindString + `",` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":null,"type":"errors:errorString","message":"err1"},` +
				`{"kind":"` + MyKindString + `","message":"err2 world","params":{"param":"world"},` +
				`"errorStack":[{"kind":"` + MyKindString + `","message":"err1 hello","params":{"param":"hello"}}]}]}`,
		},
		{
			Name: "with nested error group",

			Header:       header,
			NestedErrors: []error{erg.NewAs(erkErr1, erkErr2)},

			ExpectedKind: MyKindString,

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(erg.NewAs(erkErr1, erkErr2)),
			},

			ExpectedJSON: `{"kind":"` + MyKindString + `",` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":"` + MyKindString + `","message":"err1 hello","params":{"param":"hello"},` +
				`"errors":[{"kind":"` + MyKindString + `","message":"err2 world","params":{"param":"world"}}]` +
				`}]}`,
		},
		{
			Name: "when error header is not *erk.ExportedError and kind is set",

			Header:       BaseExporter{baseErkErr: header.(*erk.Error), kind: "my_kind"},
			NestedErrors: []error{simpleErr1, simpleErr2},

			ExpectedKind: "my_kind",

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(simpleErr1),
				erk.Export(simpleErr2),
			},

			ExpectedJSON: `{"kind":"my_kind",` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":null,"type":"errors:errorString","message":"err1"},` +
				`{"kind":null,"type":"errors:errorString","message":"err2"}]}`,
		},
		{
			Name: "when error header is not *erk.ExportedError and kind is empty",

			Header:       BaseExporter{baseErkErr: header.(*erk.Error), kind: ""},
			NestedErrors: []error{simpleErr1, simpleErr2},

			ExpectedKind: "",

			ExpectedExportedErrors: []erk.ExportedErkable{
				erk.Export(simpleErr1),
				erk.Export(simpleErr2),
			},

			ExpectedJSON: `{"kind":null,` +
				`"message":"my message my-val",` +
				`"params":{"val":"my-val"},` +
				`"errors":[{"kind":null,"type":"errors:errorString","message":"err1"},` +
				`{"kind":null,"type":"errors:errorString","message":"err2"}]}`,
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]

		err := erg.NewAs(entry.Header, entry.NestedErrors...)
		err = erk.WithParam(err, "val", "my-val")

		exported := erk.Export(err)
		ensure(exported.ErrorKind()).Equals(entry.ExpectedKind)
		ensure(exported.ErrorMessage()).Equals("my message my-val")
		ensure(exported.ErrorParams()).Equals(erk.Params{"val": "my-val"})

		errGroup := exported.(erg.ExportedGroupable)
		ensure(errGroup.GroupErrors()).Equals(entry.ExpectedExportedErrors)

		bytes, jsonErr := json.Marshal(exported)
		ensure(jsonErr).IsNotError()
		ensure(string(bytes)).Equals(entry.ExpectedJSON)
	})
}

func TestGroupAppend(t *testing.T) {
	t.Run("with no initial errors", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{.val}}:"
		err := erg.New(MyKind{}, msg)

		errs := []error{errors.New("err1"), errors.New("err2")}
		err = erg.Append(err, append(errs, nil)...)

		err = erk.WithParam(err, "val", "my-val")

		err3 := errors.New("err3")
		errs = append(errs, err3)
		err = erg.Append(err, nil, err3)

		is.Equal(erg.GetErrors(err), errs)
		is.Equal(err.Error(), "my message my-val:\n - err1\n - err2\n - err3")
	})

	t.Run("with two initial errors", func(t *testing.T) {
		is := is.New(t)

		msg := "my message {{.val}}:"
		errs := []error{errors.New("err1"), errors.New("err2")}
		err := erg.New(MyKind{}, msg, append(errs, nil)...)

		err3 := errors.New("err3")
		err4 := errors.New("err4")
		errs = append(errs, err3, err4)
		err = erg.Append(err, err3, nil, err4)

		err = erk.WithParam(err, "val", "my-val")

		err5 := errors.New("err5")
		errs = append(errs, err5)
		err = erg.Append(err, err5, nil)

		is.Equal(erg.GetErrors(err), errs)
		is.Equal(err.Error(), "my message my-val:\n - err1\n - err2\n - err3\n - err4\n - err5")
	})
}

func TestGroupMarshalJSON(t *testing.T) {
	is := is.New(t)

	group := erg.New(MyKind{}, "my group",
		erk.New(MyKind{}, "error"),
	)

	bytes, err := json.Marshal(group)
	is.NoErr(err)

	is.Equal(string(bytes),
		`{"kind":"`+MyKindString+`","message":"my group",`+
			`"errors":[{"kind":"`+MyKindString+`","message":"error"}]}`,
	)
}

type (
	baseErkErr   = erk.Error // Alias so we don't shadow the Error method
	BaseExporter struct {
		*baseErkErr
		kind string
	}
)

// Shadow Export so we can change the exported type.
func (err BaseExporter) Export() erk.ExportedErkable {
	return &erk.BaseExport{
		Kind:    err.kind,
		Message: err.Error(),
		Params:  err.Params(),
	}
}

// Shadow WithParams so we can rewrap the error.
func (err BaseExporter) WithParams(params erk.Params) error {
	errWithParams := err.baseErkErr.WithParams(params).(*erk.Error)
	return BaseExporter{baseErkErr: errWithParams, kind: err.kind}
}
