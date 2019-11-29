package erg_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/matryer/is"
)

type MyKind erk.DefaultKind

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
}

func TestErrorsString(t *testing.T) {
	t.Run("with no indentation", func(t *testing.T) {
		is := is.New(t)

		msg := "my message"
		errs := []error{errors.New("err1"), errors.New("err2")}
		err := erg.New(MyKind{}, msg, errs...)
		is.Equal(err.(erg.Groupable).ErrorsString(""), "my message:\n - err1\n - err2")
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
	is.Equal(erk.GetParams(err), nil) // Original group is not modified
}

func TestGroupKind(t *testing.T) {
	is := is.New(t)

	err := erg.New(MyKind{}, "my message")
	is.Equal(erk.GetKind(err), MyKind{})
}

func TestGroupExport(t *testing.T) {
	is := is.New(t)

	errs := []error{errors.New("err1"), errors.New("err2")}
	err := erg.New(MyKind{}, "my message {{.val}}", errs...)
	err = erk.WithParam(err, "val", "my-val")

	exported := erk.Export(err)
	is.Equal(exported.ErrorKind(), "github.com/JosiahWitt/erk/erg_test:MyKind")
	is.Equal(exported.ErrorMessage(), "my message my-val:\n - err1\n - err2")
	is.Equal(exported.ErrorParams(), erk.Params{"val": "my-val"})

	errGroup, ok := exported.(erg.ExportedGroupable)
	is.True(ok)
	is.Equal(errGroup.GroupHeader(), "my message my-val")
	is.Equal(errGroup.GroupErrors(), []string{"err1", "err2"})

	bytes, jsonErr := json.Marshal(exported)
	is.NoErr(jsonErr)
	is.Equal(string(bytes),
		`{"kind":"github.com/JosiahWitt/erk/erg_test:MyKind",`+
			`"message":"my message my-val:\n - err1\n - err2",`+
			`"params":{"val":"my-val"},`+
			`"header":"my message my-val",`+
			`"errors":["err1","err2"]}`,
	)
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
