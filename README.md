# Expression Builder 🔨

[![Go Reference](https://pkg.go.dev/badge/github.com/benpate/exp-builder.svg)](https://pkg.go.dev/github.com/benpate/exp-builder)
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

## Time Ranges

A `Time` parameter accepts a timestamp, or the name of a relative range that expands to a half-open `[begin, end)` pair anchored to midnight UTC:

```
?date=today          ?date=yesterday      ?date=tomorrow
?date=this-week      ?date=next-week
?date=last-month     ?date=this-month     ?date=next-month
?date=last-year      ?date=this-year      ?date=next-year
?date=past-30-days   ?date=next-90-days   ...and the other 30/60/90/180/365 variants
```

## What Doesn't Reach the Database

Anything you didn't ask for. A parameter that was never registered on the Builder is ignored, a value that won't parse as its declared data type is dropped, and an empty value never becomes a predicate — a `CONTAINS ""` comparison would match every record, so a blank parameter can't be used to widen a query.

That guarantee covers the *shape* of the query, not who is allowed to run it. Combine the result with your own access rules:

```go
criteria := exp.And(
	b.Evaluate(r.URL.Query()),
	exp.Equal("ownerId", currentUser.ID),
)
```

## Pull Requests Welcome

This library is a work in progress, and will benefit from your experience reports, use cases, and contributions. If you have an idea for making this library better, send in a pull request. We're all in this together! 🔨
