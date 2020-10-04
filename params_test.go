package erk_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

var _ erk.Paramable = &TestParamable{}

type TestParamable struct {
	params erk.Params
}

func (p *TestParamable) Error() string {
	return fmt.Sprint(p.params)
}

func (p *TestParamable) WithParams(params erk.Params) error {
	p.params = params
	return p
}

func (p *TestParamable) Params() erk.Params {
	return p.params
}

func TestWithParams(t *testing.T) {
	t.Run("with erk.Paramable", func(t *testing.T) {
		t.Run("setting empty params", func(t *testing.T) {
			is := is.New(t)

			err1 := &TestParamable{}
			err2 := erk.WithParams(err1, erk.Params{})
			is.Equal(err1.params, nil)
			is.Equal(err2, err1)
		})

		t.Run("setting two params", func(t *testing.T) {
			is := is.New(t)

			err1 := &TestParamable{}
			err2 := erk.WithParams(err1, erk.Params{"a": "hello", "b": "world"})
			is.Equal(err1.params, erk.Params{"a": "hello", "b": "world"})
			is.Equal(err2, err1)
		})
	})

	t.Run("with erk.Error", func(t *testing.T) {
		t.Run("setting two params", func(t *testing.T) {
			is := is.New(t)

			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err = erk.WithParams(err, erk.Params{"a": "hello", "b": "world"})
			is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
		})
	})

	t.Run("with non erk.Paramable", func(t *testing.T) {
		t.Run("setting two params", func(t *testing.T) {
			is := is.New(t)

			err1 := errors.New("hi")
			err2 := erk.WithParams(err1, erk.Params{"a": "hello", "b": "world"})
			is.Equal(erk.GetParams(err2), nil)
		})
	})
}

func TestWithParam(t *testing.T) {
	t.Run("with erk.Paramable", func(t *testing.T) {
		is := is.New(t)

		err1 := &TestParamable{}
		err2 := erk.WithParam(err1, "a", "hello")
		is.Equal(err1.params, erk.Params{"a": "hello"})
		is.Equal(err2, err1)
	})

	t.Run("with erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = erk.WithParam(err, "a", "hello")
		err = erk.WithParam(err, "b", "world")
		is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
	})

	t.Run("with non erk.Paramable", func(t *testing.T) {
		is := is.New(t)

		err1 := errors.New("hi")
		err2 := erk.WithParam(err1, "a", "hello")
		err2 = erk.WithParam(err2, "b", "world")
		is.Equal(erk.GetParams(err2), nil)
	})
}

func TestGetParams(t *testing.T) {
	t.Run("with erk.Paramable", func(t *testing.T) {
		is := is.New(t)

		err := &TestParamable{params: erk.Params{"0": "hey", "1": "there"}}
		is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("with erk.Error", func(t *testing.T) {
		is := is.New(t)

		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		is.Equal(erk.GetParams(err), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("with non erk.Paramable", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("hi")
		is.Equal(erk.GetParams(err), nil)
	})
}

func TestParamsClone(t *testing.T) {
	t.Run("returns parameters", func(t *testing.T) {
		is := is.New(t)

		params := erk.Params{"0": "hey", "1": "there"}
		is.Equal(params.Clone(), erk.Params{"0": "hey", "1": "there"})
	})

	t.Run("returns a copy of the parameters", func(t *testing.T) {
		is := is.New(t)

		params := erk.Params{"0": "hey", "1": "there"}
		paramsCopy := params.Clone()
		paramsCopy["0"] = "changed"
		is.Equal(params.Clone(), erk.Params{"0": "hey", "1": "there"})
	})
}

func TestParamsMarshalJSON(t *testing.T) {
	t.Run("with no 'err' element", func(t *testing.T) {
		is := is.New(t)

		params := erk.Params{"0": "hey", "1": "there"}
		bytes, err := json.Marshal(params)
		is.NoErr(err)
		is.Equal(string(bytes), `{"0":"hey","1":"there"}`)
	})

	t.Run("with 'err' element that is not an error", func(t *testing.T) {
		is := is.New(t)

		params := erk.Params{"0": "hey", "err": "there"}
		bytes, err := json.Marshal(params)
		is.NoErr(err)
		is.Equal(string(bytes), `{"0":"hey","err":"there"}`)
	})

	t.Run("with 'err' element that is an error", func(t *testing.T) {
		is := is.New(t)

		params := erk.Params{"0": "hey", "err": errors.New("my error")}
		bytes, err := json.Marshal(params)
		is.NoErr(err)
		is.Equal(string(bytes), `{"0":"hey","err":"my error"}`)
	})
}
