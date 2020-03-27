package erk_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/matryer/is"
)

type TestKindable struct {
	kind erk.Kind
}

func (k *TestKindable) Kind() erk.Kind {
	return k.kind
}

func (k *TestKindable) Error() string {
	return fmt.Sprintf("%T", k.kind)
}

type TestKindStringFor struct{ erk.DefaultKind }

func (TestKindStringFor) KindStringFor(k erk.Kind) string {
	return "my_kind"
}

func TestIsKind(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		t.Run("with equal kind", func(t *testing.T) {
			is := is.New(t)

			err := &TestKindable{kind: ErkExample{}}
			is.True(erk.IsKind(err, ErkExample{}))
		})

		t.Run("with non equal kind", func(t *testing.T) {
			is := is.New(t)

			err := &TestKindable{kind: ErkExample{}}
			is.Equal(erk.IsKind(err, ErkExample2{}), false)
		})
	})

	t.Run("with erk.Error", func(t *testing.T) {
		t.Run("with equal kind", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			is.True(erk.IsKind(err, ErkExample{}))
		})

		t.Run("with non equal kind", func(t *testing.T) {
			is := is.New(t)

			err := erk.New(ErkExample{}, "my message")
			is.Equal(erk.IsKind(err, ErkExample2{}), false)
		})
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		t.Run("with not equal kind", func(t *testing.T) {
			is := is.New(t)

			err := errors.New("abc")
			is.Equal(erk.IsKind(err, ErkExample{}), false)
		})

		t.Run("with equal kind", func(t *testing.T) {
			is := is.New(t)

			err := errors.New("abc")
			is.True(erk.IsKind(err, nil))
		})
	})
}

func TestGetKind(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: ErkExample{}}
		is.Equal(erk.GetKind(err), ErkExample{})
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKind(err), nil)
	})
}

func TestGetKindString(t *testing.T) {
	t.Run("with erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: ErkExample{}}
		is.Equal(erk.GetKindString(err), "github.com/JosiahWitt/erk_test:ErkExample")
	})

	t.Run("with erk.KindStringFor", func(t *testing.T) {
		is := is.New(t)

		err := &TestKindable{kind: TestKindStringFor{}}
		is.Equal(erk.GetKindString(err), "my_kind")
	})

	t.Run("with non erk.Kindable", func(t *testing.T) {
		is := is.New(t)

		err := errors.New("abc")
		is.Equal(erk.GetKindString(err), "")
	})
}

func TestKindTemplateFuncsFor(t *testing.T) {
	is := is.New(t)

	funcMap := erk.DefaultKind{}.TemplateFuncsFor(erk.DefaultKind{})
	funcMap["abc"] = func() string { return "hey" }

	funcMap2 := erk.DefaultKind{}.TemplateFuncsFor(erk.DefaultKind{})
	_, ok := funcMap2["abc"]
	is.True(!ok) // Returned func map should be a copy
}

func TestDefaultKindCloneKind(t *testing.T) {
	type Entry struct {
		Name         string
		Kind         erk.Kind
		ExpectedKind erk.Kind
		CloneCheck   func(is *is.I, entry *Entry, kindCopy erk.Kind)
	}

	table := []Entry{
		{
			Name:         "with non pointer",
			Kind:         KindWithField{Field: "hey"},
			ExpectedKind: KindWithField{Field: "hey"},
			CloneCheck: func(is *is.I, entry *Entry, kindCopyRaw erk.Kind) {
				kindCopy, ok := kindCopyRaw.(KindWithField)
				is.True(ok)

				kindCopy.Field = "something else"
				is.Equal(entry.Kind, entry.ExpectedKind)
			},
		},
		{
			Name:         "with pointer to struct",
			Kind:         &KindWithField{Field: "hey"},
			ExpectedKind: &KindWithField{Field: "hey"},
			CloneCheck: func(is *is.I, entry *Entry, kindCopyRaw erk.Kind) {
				kindCopy, ok := kindCopyRaw.(*KindWithField)
				is.True(ok)

				kindCopy.Field = "something else"
				is.Equal(entry.Kind, entry.ExpectedKind)
			},
		},
		{
			Name:         "with pointer to struct with a pointer field",
			Kind:         &KindWithPointerField{Field: PointerField("hey")},
			ExpectedKind: &KindWithPointerField{Field: PointerField("hey")},
			CloneCheck: func(is *is.I, entry *Entry, kindCopyRaw erk.Kind) {
				kindCopy, ok := kindCopyRaw.(*KindWithPointerField)
				is.True(ok)

				kindCopy.Field = PointerField("something else")
				is.Equal(entry.Kind, entry.ExpectedKind)
			},
		},
		{
			Name:         "with pointer to non struct",
			Kind:         NewKindAsStringPtr("hey"),
			ExpectedKind: NewKindAsStringPtr("hey"),
			CloneCheck: func(is *is.I, entry *Entry, kindCopyRaw erk.Kind) {
				kindCopy, ok := kindCopyRaw.(*KindAsString)
				is.True(ok)

				// This is a case we may want to eventually support
				*kindCopy = "something else"
				is.Equal(entry.Kind, NewKindAsStringPtr("something else")) // It changes the original kind
			},
		},
	}

	for _, entry := range table {
		entry := entry // Pin range variable

		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)

			kindCopy := erk.DefaultKind{}.CloneKind(entry.Kind)
			is.Equal(kindCopy, entry.ExpectedKind)
			is.Equal(entry.Kind, entry.ExpectedKind)

			if entry.CloneCheck != nil {
				entry.CloneCheck(is, &entry, kindCopy)
			}
		})
	}
}

type KindWithField struct {
	erk.DefaultKind
	Field string
}

type KindWithPointerField struct {
	erk.DefaultKind
	Field *string
}

func PointerField(str string) *string {
	return &str
}

type KindAsString string

func (k KindAsString) KindStringFor(erk.Kind) string {
	return string(k)
}

func (k KindAsString) String() string {
	return string(k)
}

func NewKindAsStringPtr(str string) *KindAsString {
	k := KindAsString(str)
	return &k
}
