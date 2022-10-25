package erk_test

import (
	"errors"
	"fmt"
	"testing"
	"text/template"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
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
	ensure := ensure.New(t)

	ensure.Run("with erk.Kindable", func(ensure ensurepkg.Ensure) {
		ensure.Run("with equal kind", func(ensure ensurepkg.Ensure) {
			err := &TestKindable{kind: ErkExample{}}
			ensure(erk.IsKind(err, ErkExample{})).IsTrue()
		})

		ensure.Run("with non equal kind", func(ensure ensurepkg.Ensure) {
			err := &TestKindable{kind: ErkExample{}}
			ensure(erk.IsKind(err, ErkExample2{})).IsFalse()
		})
	})

	ensure.Run("with erk.Error", func(ensure ensurepkg.Ensure) {
		ensure.Run("with equal kind", func(ensure ensurepkg.Ensure) {
			err := erk.New(ErkExample{}, "my message")
			ensure(erk.IsKind(err, ErkExample{})).IsTrue()
		})

		ensure.Run("with non equal kind", func(ensure ensurepkg.Ensure) {
			err := erk.New(ErkExample{}, "my message")
			ensure(erk.IsKind(err, ErkExample2{})).IsFalse()
		})
	})

	ensure.Run("with non erk.Kindable", func(ensure ensurepkg.Ensure) {
		ensure.Run("with not equal kind", func(ensure ensurepkg.Ensure) {
			err := errors.New("abc")
			ensure(erk.IsKind(err, ErkExample{})).IsFalse()
		})

		ensure.Run("with equal kind", func(ensure ensurepkg.Ensure) {
			err := errors.New("abc")
			ensure(erk.IsKind(err, nil)).IsTrue()
		})
	})
}

func TestGetKind(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Kindable", func(ensure ensurepkg.Ensure) {
		err := &TestKindable{kind: ErkExample{}}
		ensure(erk.GetKind(err)).Equals(ErkExample{})
	})

	ensure.Run("with non erk.Kindable", func(ensure ensurepkg.Ensure) {
		err := errors.New("abc")
		ensure(erk.GetKind(err)).IsNil()
	})
}

func TestGetKindString(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk.Kindable", func(ensure ensurepkg.Ensure) {
		err := &TestKindable{kind: ErkExample{}}
		ensure(erk.GetKindString(err)).Equals("github.com/JosiahWitt/erk_test:ErkExample")
	})

	ensure.Run("with erk.KindStringFor", func(ensure ensurepkg.Ensure) {
		err := &TestKindable{kind: TestKindStringFor{}}
		ensure(erk.GetKindString(err)).Equals("my_kind")
	})

	ensure.Run("with non erk.Kindable", func(ensure ensurepkg.Ensure) {
		err := errors.New("abc")
		ensure(erk.GetKindString(err)).IsEmpty()
	})
}

func TestTemplateFuncsForMethods(t *testing.T) {
	ensure := ensure.New(t)

	type templateFuncsForKind interface {
		TemplateFuncsFor(kind erk.Kind) template.FuncMap
	}

	testWithKind := func(baseKind templateFuncsForKind) {
		ensure.Run(fmt.Sprintf("with kind: %T", baseKind), func(ensurepkg.Ensure) {
			funcMap := baseKind.TemplateFuncsFor(ErkExample{})
			funcMap["abc"] = func() string { return "hey" }

			funcMap2 := baseKind.TemplateFuncsFor(ErkExample{})
			_, ok := funcMap2["abc"]
			ensure(ok).IsFalse() // Returned func map should be a copy
		})
	}

	testWithKind(erk.DefaultKind{})
	testWithKind(&erk.DefaultPtrKind{})
}

func TestKindStringForMethods(t *testing.T) {
	ensure := ensure.New(t)

	testWithKind := func(baseKind erk.Kind) {
		ensure.Run(fmt.Sprintf("with kind: %T", baseKind), func(ensure ensurepkg.Ensure) {
			ensure.Run("with value kind", func(ensure ensurepkg.Ensure) {
				kindString := baseKind.KindStringFor(ErkExample{})
				ensure(kindString).Equals("github.com/JosiahWitt/erk_test:ErkExample")
			})

			ensure.Run("with pointer kind", func(ensure ensurepkg.Ensure) {
				kindString := baseKind.KindStringFor(&ErkExample{})
				ensure(kindString).Equals("github.com/JosiahWitt/erk_test:ErkExample")
			})
		})
	}

	testWithKind(erk.DefaultKind{})
	testWithKind(&erk.DefaultPtrKind{})
}

func TestCloneKindMethods(t *testing.T) {
	ensure := ensure.New(t)

	type Entry struct {
		Name         string
		Kind         erk.Kind
		ExpectedKind erk.Kind
		CloneCheck   func(ensure ensurepkg.Ensure, entry *Entry, kindCopy erk.Kind)
	}

	testWithKind := func(baseKind interface{ CloneKind(erk.Kind) erk.Kind }) {
		ensure.Run(fmt.Sprintf("with kind: %T", baseKind), func(ensure ensurepkg.Ensure) {
			table := []Entry{
				{
					Name:         "with non pointer",
					Kind:         KindWithField{Field: "hey"},
					ExpectedKind: KindWithField{Field: "hey"},
					CloneCheck: func(ensure ensurepkg.Ensure, entry *Entry, kindCopyRaw erk.Kind) {
						kindCopy, ok := kindCopyRaw.(KindWithField)
						ensure(ok).IsTrue()

						kindCopy.Field = "something else"
						ensure(entry.Kind).Equals(entry.ExpectedKind)
					},
				},
				{
					Name:         "with pointer to struct",
					Kind:         &KindWithField{Field: "hey"},
					ExpectedKind: &KindWithField{Field: "hey"},
					CloneCheck: func(ensure ensurepkg.Ensure, entry *Entry, kindCopyRaw erk.Kind) {
						kindCopy, ok := kindCopyRaw.(*KindWithField)
						ensure(ok).IsTrue()

						kindCopy.Field = "something else"
						ensure(entry.Kind).Equals(entry.ExpectedKind)
					},
				},
				{
					Name:         "with pointer to struct with a pointer field",
					Kind:         &KindWithPointerField{Field: PointerField("hey")},
					ExpectedKind: &KindWithPointerField{Field: PointerField("hey")},
					CloneCheck: func(ensure ensurepkg.Ensure, entry *Entry, kindCopyRaw erk.Kind) {
						kindCopy, ok := kindCopyRaw.(*KindWithPointerField)
						ensure(ok).IsTrue()

						kindCopy.Field = PointerField("something else")
						ensure(entry.Kind).Equals(entry.ExpectedKind)
					},
				},
				{
					Name:         "with pointer to non struct",
					Kind:         NewKindAsStringPtr("hey"),
					ExpectedKind: NewKindAsStringPtr("hey"),
					CloneCheck: func(ensure ensurepkg.Ensure, entry *Entry, kindCopyRaw erk.Kind) {
						kindCopy, ok := kindCopyRaw.(*KindAsString)
						ensure(ok).IsTrue()

						// This is a case we may want to eventually support
						*kindCopy = "something else"
						ensure(entry.Kind).Equals(NewKindAsStringPtr("something else")) // It changes the original kind
					},
				},
			}

			ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
				entry := table[i]

				kindCopy := baseKind.CloneKind(entry.Kind)
				ensure(kindCopy).Equals(entry.ExpectedKind)
				ensure(entry.Kind).Equals(entry.ExpectedKind)

				if entry.CloneCheck != nil {
					entry.CloneCheck(ensure, &entry, kindCopy)
				}
			})
		})
	}

	testWithKind(erk.DefaultKind{})
	testWithKind(&erk.DefaultPtrKind{})
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
