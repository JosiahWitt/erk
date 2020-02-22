# erk
Errors with kinds for Go 1.13+.

[![GoDoc](https://godoc.org/github.com/JosiahWitt/erk?status.svg)](https://godoc.org/github.com/JosiahWitt/erk)
[![CI](https://github.com/JosiahWitt/erk/workflows/CI/badge.svg)](https://github.com/JosiahWitt/erk/actions?query=branch%3Amaster+workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/JosiahWitt/erk)](https://goreportcard.com/report/github.com/JosiahWitt/erk)


## Install
```
go get github.com/JosiahWitt/erk
```


## About
Erk allows you to create errors that have a kind, message template, and params.

It is recommended to define your error kind types in each package.
This allows `erk.Export` or `erk.GetKindString` to contain which package the error kind was defined, and therefore, where the error originated.


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
