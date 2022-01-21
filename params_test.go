package erk_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
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
	ensure := ensure.New(t)

	ensure.Run("with erk.Paramable", func(ensure ensurepkg.Ensure) {
		ensure.Run("setting empty params", func(ensure ensurepkg.Ensure) {
			err1 := &TestParamable{}
			err2 := erk.WithParams(err1, erk.Params{})
			ensure(err1.params).IsEmpty()
			ensure(err2).Equals(err1)
		})

		ensure.Run("setting two params", func(ensure ensurepkg.Ensure) {
			err1 := &TestParamable{}
			err2 := erk.WithParams(err1, erk.Params{"a": "hello", "b": "world"})
			ensure(err1.params).Equals(erk.Params{"a": "hello", "b": "world"})
			ensure(err2).Equals(err1)
		})
	})

	ensure.Run("with erk.Error", func(ensure ensurepkg.Ensure) {
		ensure.Run("setting two params", func(ensure ensurepkg.Ensure) {
			err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
			err = erk.WithParams(err, erk.Params{"a": "hello", "b": "world"})
			ensure(erk.GetParams(err)).Equals(erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
		})
	})

	ensure.Run("with non erk.Paramable", func(ensure ensurepkg.Ensure) {
		ensure.Run("setting two params", func(ensure ensurepkg.Ensure) {
			err1 := errors.New("hi")
			err2 := erk.WithParams(err1, erk.Params{"a": "hello", "b": "world"})
			ensure(erk.GetParams(err2)).IsEmpty()
		})
	})
}

func TestWithParam(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Paramable", func(ensure ensurepkg.Ensure) {
		err1 := &TestParamable{}
		err2 := erk.WithParam(err1, "a", "hello")
		ensure(err1.params).Equals(erk.Params{"a": "hello"})
		ensure(err2).Equals(err1)
	})

	ensure.Run("with erk.Error", func(ensure ensurepkg.Ensure) {
		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		err = erk.WithParam(err, "a", "hello")
		err = erk.WithParam(err, "b", "world")
		ensure(erk.GetParams(err)).Equals(erk.Params{"0": "hey", "1": "there", "a": "hello", "b": "world"})
	})

	ensure.Run("with non erk.Paramable", func(ensure ensurepkg.Ensure) {
		err1 := errors.New("hi")
		err2 := erk.WithParam(err1, "a", "hello")
		err2 = erk.WithParam(err2, "b", "world")
		ensure(erk.GetParams(err2)).IsEmpty()
	})
}

func TestGetParams(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Paramable", func(ensure ensurepkg.Ensure) {
		err := &TestParamable{params: erk.Params{"0": "hey", "1": "there"}}
		ensure(erk.GetParams(err)).Equals(erk.Params{"0": "hey", "1": "there"})
	})

	ensure.Run("with erk.Error", func(ensure ensurepkg.Ensure) {
		err := erk.NewWith(ErkExample{}, "my message", erk.Params{"0": "hey", "1": "there"})
		ensure(erk.GetParams(err)).Equals(erk.Params{"0": "hey", "1": "there"})
	})

	ensure.Run("with non erk.Paramable", func(ensure ensurepkg.Ensure) {
		err := errors.New("hi")
		ensure(erk.GetParams(err)).IsEmpty()
	})
}

func TestParamsClone(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("returns parameters", func(ensure ensurepkg.Ensure) {
		params := erk.Params{"0": "hey", "1": "there"}
		ensure(params.Clone()).Equals(erk.Params{"0": "hey", "1": "there"})
	})

	ensure.Run("returns a copy of the parameters", func(ensure ensurepkg.Ensure) {
		params := erk.Params{"0": "hey", "1": "there"}
		paramsCopy := params.Clone()
		paramsCopy["0"] = "changed"
		ensure(params.Clone()).Equals(erk.Params{"0": "hey", "1": "there"})
	})

	ensure.Run("returns empty params when params are nil", func(ensure ensurepkg.Ensure) {
		var params erk.Params // Nil params
		ensure(params.Clone()).Equals(erk.Params{})
	})
}

func TestParamsMarshalJSON(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with no 'err' element", func(ensure ensurepkg.Ensure) {
		params := erk.Params{"0": "hey", "1": "there"}
		bytes, err := json.Marshal(params)
		ensure(err).IsNotError()
		ensure(string(bytes)).Equals(`{"0":"hey","1":"there"}`)
	})

	ensure.Run("with 'err' element that is not an error", func(ensure ensurepkg.Ensure) {
		params := erk.Params{"0": "hey", "err": "there"}
		bytes, err := json.Marshal(params)
		ensure(err).IsNotError()
		ensure(string(bytes)).Equals(`{"0":"hey","err":"there"}`)
	})

	ensure.Run("with 'err' element that is an error", func(ensure ensurepkg.Ensure) {
		params := erk.Params{"0": "hey", "err": errors.New("my error")}
		bytes, err := json.Marshal(params)
		ensure(err).IsNotError()
		ensure(string(bytes)).Equals(`{"0":"hey","err":"my error"}`)
	})
}
