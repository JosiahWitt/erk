package erg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/JosiahWitt/erk"
)

var (
	_ erk.Erkable         = &Group{}
	_ Groupable           = &Group{}
	_ erk.ErrorIndentable = &Group{}
)

var (
	_ erk.ExportedErkable = &ExportedGroup{}
	_ ExportedGroupable   = &ExportedGroup{}
)

// Group of errors.
type Group struct {
	header error
	errors []error
}

// ExportedGroup that can be used outside the erg package.
// A common use case is marshalling the error to JSON.
//
// ExportedGroup satisfies the erk.ExportedErkable and the erg.ExportedGroupable interface.
type ExportedGroup struct {
	*erk.ExportedError
	Errors []erk.ExportedErkable `json:"errors"`
}

// New creates an error group with a kind and message.
func New(kind erk.Kind, message string, errs ...error) error {
	return NewAs(erk.New(kind, message), errs...)
}

// NewAs creates an error group given a header error.
//
// Best combined with erk.New().
func NewAs(header error, errs ...error) error {
	g := &Group{header: header}
	return g.Append(errs...)
}

// Header of the error group.
func (g *Group) Header() error {
	return g.header
}

// Error implements the error interface.
// It prints the header and list of errors.
func (g *Group) Error() string {
	return g.IndentError(" ")
}

// IndentError converts the error group to a string given the provided indentation.
func (g *Group) IndentError(indentLevel string) string {
	if indentLevel == "" {
		indentLevel = " "
	}

	str := g.header.Error()

	if !strings.HasSuffix(str, ":") && len(g.errors) > 0 {
		str += ":"
	}

	for _, err := range g.errors {
		str += fmt.Sprintf("\n%s- %s", indentLevel, buildIndentedErrorMessage(err, indentLevel))
	}

	return str
}

func buildIndentedErrorMessage(err error, indentLevel string) string {
	if indentable, ok := err.(erk.ErrorIndentable); ok {
		return indentable.IndentError(indentLevel + erk.IndentSpaces) // Add indentation to each level
	}

	return err.Error()
}

// Is implements the Go 1.13+ Is interface for use with errors.Is.
//
// Is first checks for a match against the group header,
// and then checks for a match against each error in the group.
func (g *Group) Is(target error) bool {
	if errors.Is(g.header, target) {
		return true
	}

	for _, err := range g.errors {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

// WithParams adds params to the group header.
func (g *Group) WithParams(params erk.Params) error {
	g2 := g.clone()
	g2.header = erk.WithParams(g.header, params)
	return g2
}

// Params gets params from the group header.
func (g *Group) Params() erk.Params {
	return erk.GetParams(g.header)
}

// Kind returns the error Kind of the group header.
func (g *Group) Kind() erk.Kind {
	return erk.GetKind(g.header)
}

// ExportRawMessage without executing the template.
func (g *Group) ExportRawMessage() string {
	if erkable, ok := g.header.(erk.Erkable); ok {
		return erkable.ExportRawMessage()
	}

	return g.header.Error()
}

// Export the group to an ExportedGroup.
func (g *Group) Export() erk.ExportedErkable {
	exportedErrs := []erk.ExportedErkable{}
	for _, err := range g.errors {
		exportedErrs = append(exportedErrs, erk.Export(err))
	}

	return &ExportedGroup{
		ExportedError: g.buildExportedHeader(),
		Errors:        exportedErrs,
	}
}

// Append errors to the group.
// Skips nil errors.
func (g *Group) Append(errs ...error) error {
	g2 := g.clone()
	for _, err := range errs {
		if err != nil {
			g2.errors = append(g2.errors, err)
		}
	}

	return g2
}

// Errors returns a copy of all errors of the group.
func (g *Group) Errors() []error {
	return g.clone().errors
}

// MarshalJSON by exporting the error group and then marshalling.
func (g *Group) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Export())
}

func (g *Group) clone() *Group {
	errorsCopy := make([]error, len(g.errors))
	copy(errorsCopy, g.errors)

	return &Group{
		header: g.header,
		errors: errorsCopy,
	}
}

func (g *Group) buildExportedHeader() *erk.ExportedError {
	exportedHeader := erk.Export(g.header)

	if asExportedError, ok := exportedHeader.(*erk.ExportedError); ok {
		return asExportedError
	}

	return &erk.ExportedError{
		Type:    nil,
		Kind:    toNullableKind(exportedHeader.ErrorKind()),
		Message: exportedHeader.ErrorMessage(),
		Params:  exportedHeader.ErrorParams(),

		ErrorStack: nil,
	}
}

func toNullableKind(kind string) *string {
	if kind == "" {
		return nil
	}

	return &kind
}

// GroupErrors satisfies the ExportedGroupable interface.
func (g *ExportedGroup) GroupErrors() []erk.ExportedErkable {
	return g.Errors
}
