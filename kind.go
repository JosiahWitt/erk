package erk

import (
	"errors"
	"fmt"
	"reflect"
	"text/template"
)

// Kind represents an error kind.
//
// Example:
//
//	package hello
//
//	type (
//	  ErkJSONUnmarshalling struct { erk.DefaultKind }
//	  ErkJSONMarshalling   struct { erk.DefaultKind }
//	)
//
//	...
//
//	// Creating an error with the kind
//	err := erk.New(ErkJSONUnmarshalling, "failed to unmarshal JSON: '{{.json}}'") // Usually this would be a global error variable
//	err = erk.WithParams(err, "json", originalJSON)
//
//	...
type Kind interface {
	KindStringFor(Kind) string
}

// DefaultKind should be embedded in most Kinds.
//
// It is recommended to create new error kinds in each package.
// This allows erk to get the package name the error occurred in.
//
// Example: See Kind.
type DefaultKind struct{}

// DefaultKind implements Kind.
var _ Kind = DefaultKind{}

// DefaultPtrKind is equivalent to DefaultKind, but enforces that the kinds are pointers.
type DefaultPtrKind struct{}

// DefaultPtrKind implements Kind.
var _ Kind = &DefaultPtrKind{}

// Kindable errors that support housing an error Kind.
type Kindable interface {
	Kind() Kind
}

// IsKind checks if the error's kind is the provided kind.
func IsKind(err error, kind Kind) bool {
	return reflect.TypeOf(GetKind(err)) == reflect.TypeOf(kind)
}

// GetKind from the provided error.
func GetKind(err error) Kind {
	var k Kindable
	if errors.As(err, &k) {
		return k.Kind()
	}

	return nil
}

// GetKindString returns a string identifying the kind of the error.
//
// If the kind embeds erk.DefaultKind, this will be a string with the package and type of the error's kind.
// This string can be overridden by implementing a KindStringFor method on a base kind, and embedding that in the error kind.
//
// erk.DefaultKind Example:
//
//	erk.GetKindString(err) // Output: "github.com/username/package:ErkYourKind"
func GetKindString(err error) string {
	k := GetKind(err)
	if k == nil {
		return ""
	}

	return k.KindStringFor(k)
}

func cloneKind(kind Kind) Kind {
	if clonable, ok := kind.(interface{ CloneKind(Kind) Kind }); ok {
		return clonable.CloneKind(kind)
	}

	return DefaultKind{}.CloneKind(kind)
}

// KindStringFor the provided kind.
func (DefaultKind) KindStringFor(kind Kind) string {
	return buildDefaultKindString(kind)
}

// TemplateFuncsFor the provided kind.
func (DefaultKind) TemplateFuncsFor(kind Kind) template.FuncMap {
	funcMap := make(template.FuncMap, len(defaultTemplateFuncs))
	for k, v := range defaultTemplateFuncs {
		funcMap[k] = v
	}

	return funcMap
}

// CloneKind to a shallow copy.
//
// If the kind is not a pointer, it is directly returned (since it was passed by value).
// If the kind is not a struct, it is directly returned (not supported for now).
// Otherwise, a shallow new copy of the struct is created using reflection, and the first layer of the struct is copyied using Set.
func (DefaultKind) CloneKind(kind Kind) Kind {
	originalKind := reflect.ValueOf(kind)
	if originalKind.Kind() != reflect.Ptr {
		return kind
	}

	if originalKind.Elem().Kind() != reflect.Struct {
		return kind
	}

	kindCopy := reflect.New(originalKind.Elem().Type())
	kindCopy.Elem().Set(originalKind.Elem())
	return kindCopy.Interface().(Kind) //nolint:forcetypeassert // We know this implements Kind
}

// KindStringFor the provided kind.
func (*DefaultPtrKind) KindStringFor(kind Kind) string {
	return DefaultKind{}.KindStringFor(kind)
}

// TemplateFuncsFor the provided kind.
func (*DefaultPtrKind) TemplateFuncsFor(kind Kind) template.FuncMap {
	return DefaultKind{}.TemplateFuncsFor(kind)
}

// CloneKind to a shallow copy.
func (*DefaultPtrKind) CloneKind(kind Kind) Kind {
	return DefaultKind{}.CloneKind(kind)
}

func buildDefaultKindString(kind interface{}) string {
	t := reflect.TypeOf(kind)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf("%s:%s", t.PkgPath(), t.Name())
}
