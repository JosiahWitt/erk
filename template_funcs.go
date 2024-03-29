package erk

import (
	"fmt"
	"text/template"
)

func templateFuncs(k Kind) template.FuncMap {
	if funcs, ok := k.(interface{ TemplateFuncsFor(Kind) template.FuncMap }); ok {
		return funcs.TemplateFuncsFor(k)
	}

	return defaultTemplateFuncs
}

// Functions that are accessible from the error templates.
//
//nolint:gochecknoglobals // Only read internally
var defaultTemplateFuncs = template.FuncMap{
	"type":    templateFuncType,
	"inspect": templateFuncInspect,
}

func templateFuncType(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func templateFuncInspect(v interface{}) string {
	return fmt.Sprintf("%+v", v)
}
