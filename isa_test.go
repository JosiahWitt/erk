package erk_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type myIsAErrorType1 struct {
	Value string
}

func (e *myIsAErrorType1) Error() string {
	return e.Value
}

type myIsAErrorType2 struct {
	Value string
}

func (e *myIsAErrorType2) Error() string {
	return e.Value
}

type myIsAErrorType3 struct {
	Value string
}

func (e *myIsAErrorType3) Error() string {
	return e.Value
}

func (e *myIsAErrorType3) Is(target error) bool {
	return true
}

type myIsAErrorType4 struct {
	Value string
}

func (e *myIsAErrorType4) Error() string {
	return e.Value
}

func TestIsA(t *testing.T) {
	table := []struct {
		Name   string
		Source error
		Target error
		Result bool
	}{
		{
			Name:   "with errors.Is equal errors",
			Source: &myIsAErrorType3{Value: "abc"},
			Target: &myIsAErrorType4{},
			Result: true,
		},
		{
			Name:   "with errors.Is equal errors and target errors.New error",
			Source: &myIsAErrorType3{},
			Target: errors.New("same string"),
			Result: true,
		},
		{
			Name:   "with equal type errors",
			Source: &myIsAErrorType1{Value: "abc"},
			Target: &myIsAErrorType1{},
			Result: true,
		},
		{
			Name:   "with different type errors",
			Source: &myIsAErrorType1{Value: "abc"},
			Target: &myIsAErrorType2{},
			Result: false,
		},
		{
			Name:   "with two errors.New errors",
			Source: errors.New("same string"),
			Target: errors.New("same string"),
			Result: false,
		},
		{
			Name:   "with source errors.New error",
			Source: errors.New("same string"),
			Target: &myIsAErrorType4{},
			Result: false,
		},
		{
			Name:   "with target errors.New error",
			Source: &myIsAErrorType2{},
			Target: errors.New("same string"),
			Result: false,
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			result := erk.IsA(entry.Source, entry.Target)
			is.Equal(result, entry.Result)
		})
	}
}

func TestIsAStringError(t *testing.T) {
	table := []struct {
		Name   string
		Err    error
		Result bool
	}{
		{
			Name:   "with a string error",
			Err:    errors.New("my error"),
			Result: true,
		},
		{
			Name:   "with a non-string error",
			Err:    &myIsAErrorType4{},
			Result: false,
		},
	}

	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			result := erk.IsAStringError(entry.Err)
			is.Equal(result, entry.Result)
		})
	}
}
