# Expression Builder 🔨

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://pkg.go.dev/github.com/benpate/exp-builder)
[![Version](https://img.shields.io/github/v/release/benpate/exp-builder?include_prereleases&style=flat-square&color=brightgreen)](https://github.com/benpate/exp-builder/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/benpate/exp-builder/go.yml?style=flat-square)](https://github.com/benpate/exp-builder/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/benpate/exp-builder?style=flat-square)](https://goreportcard.com/report/github.com/benpate/exp-builder)
[![Codecov](https://img.shields.io/codecov/c/github/benpate/exp-builder.svg?style=flat-square)](https://codecov.io/gh/benpate/exp-builder)

## Safe queries from URL query strings

Expression Builder works with the [exp expression library](https://github.com/benpate/exp) to define templates that are safely populated with data from a URL query string.  More documentation to come.

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

## Pull Requests Welcome

This library is a work in progress, and will benefit from your experience reports, use cases, and contributions.  If you have an idea for making this library better, send in a pull request.  We're all in this together! 🔨
