# erk
Errors with kinds for Go 1.13+.

[![GoDoc](https://godoc.org/github.com/JosiahWitt/erk?status.svg)](https://godoc.org/github.com/JosiahWitt/erk)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/JosiahWitt/erk)
[![CI](https://github.com/JosiahWitt/erk/workflows/CI/badge.svg)](https://github.com/JosiahWitt/erk/actions?query=branch%3Amaster+workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/JosiahWitt/erk)](https://goreportcard.com/report/github.com/JosiahWitt/erk)
[![codecov](https://codecov.io/gh/JosiahWitt/erk/branch/master/graph/badge.svg)](https://codecov.io/gh/JosiahWitt/erk)

## Table of Contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Install](#install)
- [About](#about)
- [Overview](#overview)
  - [Error Kinds](#error-kinds)
  - [Message Templates](#message-templates)
    - [Template Functions](#template-functions)
      - [Extending Template Functions](#extending-template-functions)
  - [Params](#params)
    - [Wrapping Errors](#wrapping-errors)
  - [Error Groups](#error-groups)
  - [Testing](#testing)
    - [Mocking](#mocking)
  - [Strict Mode](#strict-mode)
  - [JSON Errors](#json-errors)
  - [Advanced Kinds](#advanced-kinds)
    - [Warnings](#warnings)
    - [HTTP Statuses](#http-statuses)
- [Recommendations](#recommendations)
  - [Default Error Kind](#default-error-kind)
  - [Defining Error Kinds](#defining-error-kinds)
  - [Defining Errors](#defining-errors)
- [Examples](#examples)
  - [Error Kinds](#error-kinds-1)
    - [Output](#output)
  - [Error Groups](#error-groups-1)
    - [Output](#output-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

---

## Install
```bash
$ go get github.com/JosiahWitt/erk
```


## About
Erk allows you to create errors that have a kind, message template, and params.

Since Erk supports Go 1.13+ [`errors.Is`](https://pkg.go.dev/errors?tab=doc#Is), it is easier to test errors, especially errors that contain parameters.

Erk is quite extensible by leveraging the fact that kinds are struct types.
For example, HTTP status codes, distinguishing between warnings and errors, and more can easily be embedded in kinds.
See [advanced kinds](#advanced-kinds) for some examples.

The name "erk" comes from "errors with kinds".
Erk is also a play on _[irk](https://www.merriam-webster.com/dictionary/irk)_, since errors can be annoying to deal with.
Hopefully Erk makes them less irksome. 😄


## Overview

### Error Kinds
Error kinds are struct types that implement the [`Kind`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Kind) interface.
Typically the `Kind` interface is satisfied by embedding a default kind, such as [`erk.DefaultKind`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#DefaultKind).
It is recommended to [define a default kind](#default-error-kind) for your app or package.
> Example: `type ErkTableMissing struct { erk.DefaultKind }`

### Message Templates
Error messages are [text templates](https://pkg.go.dev/text/template?tab=doc), which allows referencing params by name.
Since params are stored in map, this is done by using the `{{.paramName}}` notation.
> Example: `"table {{.tableName}} does not exist"`

#### Template Functions
A few functions in addition to the [built in](https://pkg.go.dev/text/template?tab=doc#hdr-Functions) template functions have been added.
- `type`: Returns the type of the param. It is equivalent to `fmt.Sprintf("%T", param)`
  > Example: `{{type .paramName}}`

- `inspect`: Returns more details for complex types. It is equivalent to `fmt.Sprintf("%+v", param)`
  > Example: `{{inspect .paramName}}`

##### Extending Template Functions
Template functions can be extended by overriding the [`TemplateFuncsFor`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#DefaultKind.TemplateFuncsFor) method on your [default kind](#default-error-kind).

### Params
Params allow adding arbitrary context to errors.
Params are stored as a map, and can be referenced in templates.

#### Wrapping Errors
Other errors can be wrapped into Erk errors using the [`erk.Wrap`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Wrap), [`erk.WrapAs`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#WrapAs), and [`erk.WrapWith`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#WrapWith), functions.
(I recommend [defining errors as public variables](#defining-errors), and avoid using `erk.Wrap`.)

The wrapped error is stored in the params by the [`err` key](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#pkg-constants).
Thus, templates can reference the error they wrap by using `{{.err}}`.

Use [`errors.Unwrap`](https://pkg.go.dev/errors?tab=doc#Unwrap) to return the original error.

### Error Groups
Errors can be grouped using the [`erg`](https://pkg.go.dev/github.com/JosiahWitt/erk/erg?tab=doc) package.

Errors are [appended](https://pkg.go.dev/github.com/JosiahWitt/erk/erg?tab=doc#Append) to the error group as they are encountered.
Be sure to conditionally return the error group by calling [`erg.Any`](https://pkg.go.dev/github.com/JosiahWitt/erk/erg?tab=doc#Any), otherwise a non-nil error group with no errors will be returned.

See [the example](#error-groups-1) below.

### Testing
Since Erk supports Go 1.13+ [`errors.Is`](https://pkg.go.dev/errors?tab=doc#Is), testing errors is straightforward.
This is especially helpful for comparing errors that leverage parameters, since the parameters are ignored.
(Usually you just want to test a certain error was returned from the function, not that the error is assembled correctly.)

> Example: `errors.Is(err, mypkg.ErrTableDoesNotExist)` returns `true` only if the `err` is `mypkg.ErrTableDoesNotExist`

#### Mocking
When returning an Erk error from a mock, most of the time the required template parameters are not critical to the test.
However, if the code being tested uses [`errors.Is`](https://pkg.go.dev/errors?tab=doc#Is), and [strict mode](#strict-mode) is enabled, simply returning the error from the mock will result in a panic.

> Example: `someMockedFunction.Returns(store.ErrItemNotFound)` might panic

Thus, the [`erkmock`](https://pkg.go.dev/github.com/JosiahWitt/erk/erkmock?tab=doc) package exists to support returning errors from mocks without setting the required parameters.
You can create a mocked error [`From`](https://pkg.go.dev/github.com/JosiahWitt/erk/erkmock?tab=doc#From) an existing Erk error, or [`For`](https://pkg.go.dev/github.com/JosiahWitt/erk/erkmock?tab=doc#For) an error kind.

> Example: `someMockedFunction.Returns(erkmock.From(store.ErrItemNotFound))` does not panic

### Strict Mode
By default, strict mode is not enabled.
Thus, if errors are encountered while rendering the error (eg. invalid template), the unrendered template is silently returned.
If parameters are missing for the template, `<no value>` is used instead.
This makes sense in production, as an unrendered template is better than returning a render error.

However, when testing or in development mode, it might be useful for these types of issues to be more visible.

Strict mode causes a panic when it encounters an invalid template or missing parameters.
It is automatically enabled in tests, and can be explicitly enabled or disabled using the `ERK_STRICT_MODE` environment variable set to `true` or `false`, respectively.
It can also be enabled or disabled programmatically by using the [`erkstrict.SetStrictMode`](https://pkg.go.dev/github.com/JosiahWitt/erk/erkstrict?tab=doc#SetStrictMode) function.

When strict mode is enabled, calls to [`errors.Is`](https://pkg.go.dev/errors?tab=doc#Is) will also attempt to render the error. This is useful in tests.

### JSON Errors
Errors created with Erk can be directly marshaled to JSON, since the [`MarshalJSON`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Error.MarshalJSON) method is present.

Internally, this calls [`erk.Export`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Export), followed by `json.Marshal`.

If you want to customize how errors are marshalled to JSON, simply write your own function that uses [`erk.Export`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Export) and modifies the exported error as necessary before marshalling JSON.

> If not all errors in your application are guaranteed to be `erk` errors, calling [`erk.Export`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Export) before marshalling to JSON will ensure each error is explicitly converted to an `erk` error.

> If you would like to export the errors as JSON, _and return the error kind as the error type_, see [`erkjson`](https://pkg.go.dev/github.com/JosiahWitt/erk/erkjson).
> Using the error kind as the exported error type is useful for something like AWS Step Functions, which allows defining retry policies based on the type of the returned error.


### Advanced Kinds
Since error kinds are struct types, they can embed other structs.
This allows quite a bit of flexibility.

#### Warnings
For example, you could create an `erkwarning` package that defines a struct with an `IsWarning() bool` method.
Then, you can use an interface to check for that method, and if the method returns `true`, log the error instead of returning it to the client.
This would work well when coupled with [`erg`](https://godoc.org/github.com/JosiahWitt/erk/erg).
Any error kind that should be a warning simply needs to embed the struct from `erkwarning`.
This allows all errors to bubble to the top, simplifying how warnings and errors are distinguished.

#### HTTP Statuses
Something similar can also be done for HTTP statuses, allowing status codes to be determined on the error kind level.

See [`erkhttp`](https://github.com/JosiahWitt/erkhttp) for an implementation.


## Recommendations
### Default Error Kind
It is recommended to define a default error kind for your app or package that embeds `erk.DefaultKind`.
Then, every error kind for your app or package can embed that default error kind.
This allows easily overriding or adding properties to the default kind.

Two recommended names for this shared package are `erks` or `errkinds`.

> Example: `type Default struct { erk.DefaultKind }`

### Defining Error Kinds
There are two recommended ways to define your kinds:
1. Define your error kind types in each package near the errors themselves.
   > This allows [`erk.Export`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#Export) or [`erk.GetKindString`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#GetKindString) to contain which package the error kind was defined, and therefore, where the error originated.

2. Define a package that contains all error kinds, and override the default error kind's [`KindStringFor`](https://pkg.go.dev/github.com/JosiahWitt/erk?tab=doc#DefaultKind.KindStringFor) method to return a snake case version of each kind's type.
   > This produces a nicer API for consumers, and allows you to move around error kinds without changing the string emitted by the API.

   > If using this method in a package, it may be a good idea to prefix with your package name to prevent collisions.

### Defining Errors
It is recommended to define every error as a public variable, so consumers of your package can check against each error.
Avoid defining errors inside of functions.


## Examples

### Error Kinds
You can create errors with kinds using the [`erk`](https://godoc.org/github.com/JosiahWitt/erk) package.

```go
package store

import "github.com/JosiahWitt/erk"

type (
  ErkMissingKey struct { erk.DefaultKind }
  ...
)

var (
  ErrMissingReadKey = erk.New(ErkMissingKey{}, "no read key specified for table '{{.tableName}}'")
  ErrMissingWriteKey = erk.New(ErkMissingKey{}, "no write key specified for table '{{.tableName}}'")
  ...
)

func Read(tableName, key string, data interface{}) error {
  ...

  if key == "" {
    return erk.WithParam(ErrMissingReadKey, "tableName", tableName)
  }

  ...
}
```

```go
package main

...

func main() {
  err := store.Read("my_table", "", nil)

  bytes, _ := json.MarshalIndent(erk.Export(err), "", "  ")
  fmt.Println(string(bytes))

  fmt.Println()
  fmt.Println("erk.IsKind(err, store.ErkMissingKey{}):  ", erk.IsKind(err, store.ErkMissingKey{}))
  fmt.Println("errors.Is(err, store.ErrMissingReadKey): ", errors.Is(err, store.ErrMissingReadKey))
  fmt.Println("errors.Is(err, store.ErrMissingWriteKey):", errors.Is(err, store.ErrMissingWriteKey))
}
```

#### Output
```json
{
  "kind": "github.com/username/repo/store:ErkMissingKey",
  "message": "no read key specified for table 'my_table'",
  "params": {
    "tableName": "my_table"
  }
}

erk.IsKind(err, store.ErkMissingKey{}):   true
errors.Is(err, store.ErrMissingReadKey):  true
errors.Is(err, store.ErrMissingWriteKey): false
```


### Error Groups
You can also wrap a group of errors using the [`erg`](https://godoc.org/github.com/JosiahWitt/erk/erg) package.


```go
package store

import "github.com/JosiahWitt/erk"

type (
  ErkMultiRead struct { erk.DefaultKind }
  ...
)

var (
  ErrUnableToMultiRead = erk.New(ErkMultiRead{}, "could not multi read from '{{.tableName}}'")
  ...
)

func MultiRead(tableName string, keys []string, data interface{}) error {
  ...

  groupErr := erg.NewAs(ErrUnableToMultiRead)
  groupErr = erk.WithParam(groupErr, "tableName", tableName)
  for _, key := range keys {
    groupErr = erg.Append(groupErr, Read(tableName, key, data))
  }
  if erg.Any(groupErr) {
    return groupErr
  }

  ...
}
```

```go
package main

...

func main() {
  err := store.MultiRead("my_table", []string{"", "my key", ""}, nil)

  bytes, _ := json.MarshalIndent(erk.Export(err), "", "  ")
  fmt.Println(string(bytes))

  fmt.Println()
  fmt.Println("erk.IsKind(err, store.ErkMultiRead{}):     ", erk.IsKind(err, store.ErkMultiRead{}))
  fmt.Println("errors.Is(err, store.ErrUnableToMultiRead):", errors.Is(err, store.ErrUnableToMultiRead))
  fmt.Println("errors.Is(err, store.ErrMissingReadKey):   ", errors.Is(err, store.ErrMissingReadKey))
  fmt.Println("errors.Is(err, store.ErrMissingWriteKey):  ", errors.Is(err, store.ErrMissingWriteKey))
}
```

#### Output
```json
{
  "kind": "github.com/username/repo/store:ErkMultiRead",
  "message": "could not multi read from 'my_table':\n - no read key specified for table 'my_table'\n - no read key specified for table 'my_table'",
  "params": {
    "tableName": "my_table"
  },
  "header": "could not multi read from 'my_table'",
  "errors": [
    "no read key specified for table 'my_table'",
    "no read key specified for table 'my_table'"
  ]
}

erk.IsKind(err, store.ErkMultiRead{}):      true
errors.Is(err, store.ErrUnableToMultiRead): true
errors.Is(err, store.ErrMissingReadKey):    false
errors.Is(err, store.ErrMissingWriteKey):   false
```
