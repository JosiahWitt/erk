# erk
Errors with kinds for Go 1.13+.

[![GoDoc](https://godoc.org/github.com/JosiahWitt/erk?status.svg)](https://godoc.org/github.com/JosiahWitt/erk)

## Install
```
go get github.com/JosiahWitt/erk
```

## About
Erk allows you to create errors that have a kind, message template, and params.

It is recommended to define your error kind types in each package.
This allows `erk.Export` or `erk.GetKindString` to contain which package the error kind was defined, and therefore, where the error originated.

## Example

```go
  package store

  import "github.com/JosiahWitt/erk"

  type (
    ErkMissingKey erk.DefaultKind
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

### Output
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
