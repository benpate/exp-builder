# Expression Builder 🔨

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://pkg.go.dev/github.com/benpate/exp-builder)
[![Version](https://img.shields.io/github/v/release/benpate/exp-builder?include_prereleases&style=flat-square&color=brightgreen)](https://github.com/benpate/exp-builder/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/benpate/exp-builder/go.yml?style=flat-square)](https://github.com/benpate/exp-builder/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/benpate/exp-builder?style=flat-square)](https://goreportcard.com/report/github.com/benpate/exp-builder)
[![Codecov](https://img.shields.io/codecov/c/github/benpate/exp-builder.svg?style=flat-square)](https://codecov.io/gh/benpate/exp-builder)

## Safe queries from URL query strings

Expression Builder works with the [exp expression library](https://github.com/benpate/exp) to define templates that are safely populated with data from a URL query string.

```go

func Handler(r *http.Request, w http.ResponseWriter) {

	// Define the URL arguments you want to allow and their types
	b := NewBuilder().
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	// Create a usable expression from the URL query string
	expression := b.Evaluate(r.URL.Query())

	// Next, safely pass it into the database, or something...
}
```

## Operators

By default each parameter is compared with `=`. A query value may override this with an `OP:value` prefix, where `OP` is one of the operator aliases recognized by [exp](https://github.com/benpate/exp) (`EQ`, `NE`, `GT`, `GTE`, `LT`, `LTE`, `CONTAINS`, `BEGINS`, case-insensitive):

```
?publishDate=gt:1000      // publishDate >  1000
?firstName=ne:John        // firstName  != "John"
?lastName=BEGINS:Mc       // lastName begins with "Mc"
```

A field's default operator can also be set in code with `WithDefaultOperator` (and the `WithDefaultOp*` shortcuts). Repeating a parameter (`?firstName=a&firstName=b`) combines the values with OR; distinct parameters combine with AND.

## What matters here

- **The `OP:value` prefix is the core feature, and it is invisible from the constructor API.** Only the first `:` splits operator from value, so `EQ:a:b:c` parses as `=` / `a:b:c`. An unrecognized operator prefix is *not* an error — the default operator is used and the whole string (prefix included) becomes the value.

- **Only allow-listed fields produce predicates.** `Evaluate` ignores any URL parameter whose name was not registered on the Builder, and silently drops values that fail to parse for their data type. This is the safety guarantee — unknown or malformed input never reaches the expression.

- **`Polygon` ignores the operator prefix and always emits `GEO-WITHIN`.** Unlike every other type, the parsed operator is discarded for polygons. The prefix *is* stripped before parsing the coordinates (so `eq:1,2,3,4` works), but the comparison operator is fixed.

- **`Evaluate` vs. `EvaluateAll`.** `Evaluate` includes only the fields present in the URL; `EvaluateAll` requires *every* registered field to be present and non-empty, returning an error otherwise.

## Pull Requests Welcome

This library is a work in progress, and will benefit from your experience reports, use cases, and contributions.  If you have an idea for making this library better, send in a pull request.  We're all in this together! 🔨
