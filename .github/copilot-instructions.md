# Erk - Errors with Kinds for Go

## Architecture Overview

Erk is a Go error library that adds structured error kinds, message templating, and parameter storage to standard Go errors while maintaining compatibility with Go 1.13+ `errors.Is` and `errors.Unwrap`.

**Core Components:**

- **erk**: Main package - error creation, wrapping, parameter management
- **erg**: Error groups - collect multiple errors under a single header error
- **erkstrict**: Strict mode for development/testing - panics on template/parameter issues
- **erkmock**: Mock errors for testing without setting required template parameters
- **erkjson**: JSON export with error kind as type (uses pointer kinds)

## General Instructions

- This is a library, so avoid breaking changes to public APIs unless absolutely necessary.
- Follow Go conventions and idiomatic patterns.
- Prioritize simplicity over complexity.

## Error Definition Pattern

**Define error kinds as struct types** embedding `erk.DefaultKind`:

```go
type ErkMissingKey struct { erk.DefaultKind }
```

**Define errors as public variables** (not inside functions):

```go
var ErrMissingReadKey = erk.New(ErkMissingKey{}, "no read key specified for table '{{.tableName}}'")
```

**Use errors with parameters:**

```go
return erk.WithParam(ErrMissingReadKey, "tableName", tableName)
```

This pattern allows consumers to use `errors.Is(err, pkg.ErrMissingReadKey)` regardless of parameter values.

## Message Templates

Error messages use Go `text/template` syntax. Parameters are referenced as `{{.paramName}}`.

**Custom template functions:**

- `{{type .param}}` - returns type of param (like `fmt.Sprintf("%T", ...)`)
- `{{inspect .param}}` - detailed output (like `fmt.Sprintf("%+v", ...)`)

**Wrapped errors** are accessible in templates via `{{.err}}` (stored in `erk.OriginalErrorParam`).

## Error Groups (erg)

Use `erg` to collect multiple errors:

```go
groupErr := erg.NewAs(ErrUnableToMultiRead)
groupErr = erk.WithParam(groupErr, "tableName", tableName)
for _, key := range keys {
  groupErr = erg.Append(groupErr, Read(tableName, key, data))
}
if erg.Any(groupErr) {  // Check if any errors exist
  return groupErr
}
```

**Important:** Always use `erg.Any()` to conditionally return - prevents returning non-nil empty error groups.

## Strict Mode

**Automatically enabled in tests** (detected via `-test.*` flag) or via `ERK_STRICT_MODE=true` env var.

**Behavior:**

- Panics on invalid templates or missing parameters (instead of silently returning raw template)
- Validates errors during `errors.Is()` calls (useful for catching issues in tests)

**To disable in tests:** Call `erkstrict.SetStrictMode(false)` in `init()` (see `error_test.go`)

## Testing Patterns

**Use `github.com/JosiahWitt/ensure`** for assertions (project's testing library):

```go
ensure := ensure.New(t)
ensure(result).Equals(expected)
ensure(value).IsTrue()
ensure.Run("nested test", func(ensure ensurepkg.Ensure) { ... })
```

**Test error types with `errors.Is()`** - ignores parameters:

```go
errors.Is(err, store.ErrMissingReadKey)  // true if same error kind + message template
```

**Mock errors in tests** using `erkmock.From()` to avoid strict mode panics:

```go
someMockedFunction.Returns(erkmock.From(store.ErrItemNotFound))
```

## Development Workflow

**Run tests:**

```bash
go test ./...                                    # All tests
go test -race -coverprofile=coverage.txt ./...   # With race detector and coverage
```

**Package naming convention:**

- Test packages use `_test` suffix (e.g., `package erk_test`)
- Enables testing public API as consumers would use it

**Go version:** Requires Go 1.13+ (uses `errors.Is`, `errors.As`, `errors.Unwrap`)

## Key Interfaces

- `Erkable`: Errors with Params, Kind, and Export capability
- `Paramable`: Errors that support `WithParams()` and `Params()`
- `Kindable`: Errors that support `Kind()`
- `Exportable`: Errors that support `Export()` for JSON marshaling

## Common Patterns

**Creating errors:** Use `erk.New()` or `erk.NewWith()` (for params at creation time)

**Wrapping errors:** Use `erk.WrapAs()` or `erk.WrapWith()` (avoid `erk.Wrap()` - see README recommendations)

**Converting any error to Erk:** Use `erk.ToErk()` - wraps non-Erk errors or returns existing Erk errors unchanged

**Exporting to JSON:** Use `erkjson.Export(err)` to create an error who's value is JSON, or marshal an `erk` error directly to JSON.
