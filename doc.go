// Package builder turns a URL query string into a database expression, using an
// allow-list of parameters that the caller declares up front.
//
// A Builder maps each permitted URL parameter to a data type, and Evaluate reads
// a url.Values against it:
//
//	b := builder.NewBuilder().
//		String("firstName").
//		Int("publishDate").
//		Bool("isPublished")
//
//	criteria := b.Evaluate(request.URL.Query())
//
// The result is an exp.Expression, ready to hand to a query. Repeating a
// parameter combines its values with OR; distinct parameters combine with AND.
//
// # Only declared parameters reach the expression
//
// This is the safety guarantee, and the reason to use this package rather than
// reading query parameters by hand. A parameter that was never declared on the
// Builder is ignored, and a value that cannot be parsed as its declared data
// type is dropped. Neither one is an error, and neither one reaches the
// database.
//
// An empty value, on the other hand, is a real comparison and is kept: a field
// that may be unset is queried as (field == "") OR (field == "VALUE"). Note the
// consequence for a Field that carries a WithFilter -- if the filter empties the
// value, the comparison becomes an empty one, and with CONTAINS that matches
// every record. Filters that can return "" deserve a second look.
//
// The guarantee covers SHAPE, not authorization. A caller is still responsible
// for restricting the query to records the visitor may see -- typically by
// combining the result with exp.And.
//
// # Operator prefixes
//
// Each parameter is compared with "=" by default, which WithDefaultOperator (and
// its WithDefaultOp shortcuts) can change per field. A query string may also
// name its own operator with an "OP:value" prefix, using any of the operator
// aliases that exp recognizes:
//
//	?publishDate=gt:1000      publishDate >  1000
//	?firstName=ne:John        firstName  != "John"
//	?lastName=BEGINS:Mc       lastName begins with "Mc"
//
// Only the FIRST colon splits the prefix from the value, so "eq:a:b:c" compares
// against "a:b:c". An unrecognized prefix is not an error -- the field's default
// operator is used, and the whole string, prefix included, becomes the value.
// That is what lets a timestamp such as "2026-01-01T09:30:00Z" pass through
// unharmed.
//
// Two data types define their own comparison and therefore discard the operator:
// a Polygon always emits GEO-WITHIN, and a Time whose value names a range (see
// below) always emits the half-open pair that bounds it. Both still strip the
// prefix before parsing the value.
//
// # Time ranges
//
// A Time parameter accepts a timestamp, or the name of a relative range, which
// expands to a half-open [begin, end) pair anchored to midnight UTC:
//
//	today          yesterday      tomorrow
//	this-week      next-week
//	last-month     this-month     next-month
//	last-year      this-year      next-year
//	past-30-days   past-60-days   past-90-days   past-180-days   past-365-days
//	next-30-days   next-60-days   next-90-days   next-180-days   next-365-days
//
// # Evaluate and EvaluateAll
//
// Evaluate uses whichever declared parameters are present. EvaluateAll requires
// every declared parameter to be present and non-empty, and returns a
// derp.BadRequest naming the first one that is missing.
package builder
