package erk

import (
	"fmt"
	"text/template"
)

// Functions that are accessible from the error templates.
var templateFuncs = template.FuncMap{
	"type": templateFuncType,
}

func templateFuncType(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
