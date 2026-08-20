# Expression Builder — Notes for AI Agents

- **The `OP:value` prefix is the core feature, and it is invisible from the constructor API.** Only the first `:` splits operator from value, so `EQ:a:b:c` parses as `=` / `a:b:c`. An unrecognized operator prefix is *not* an error — the default operator is used and the whole string (prefix included) becomes the value. That fallback is what lets an ISO timestamp, which is full of colons, survive the split.

- **Only allow-listed fields produce predicates.** `Evaluate` ignores any URL parameter whose name was not registered on the Builder, and silently drops values that fail to parse for their data type. This is the safety guarantee — unknown or malformed input never reaches the expression. It is a guarantee about *shape*, not about authorization: callers still have to `exp.And` their own access rules on top.

- **An empty value IS a predicate, deliberately.** `(field == "") || (field == "VALUE")` is how you query a field that may be unset, so `EvaluateField` keeps blank values. Only `Evaluate`'s `sliceNotEmpty` check skips a parameter, and only when *every* one of its values is blank. A guard that dropped empty values was written and removed — do not re-add it without reading BUG-131, which covers the unresolved case against it: a `WithFilter` that empties a value yields `CONTAINS ""`, which matches every record.

- **`Polygon` and named time ranges ignore the operator prefix.** Both strip the prefix before parsing (so `eq:1,2,3,4` and `gt:today` work), then discard it: a Polygon always emits `GEO-WITHIN`, and a named range always emits its own `>= begin AND < end` pair. Every other data type honors the parsed operator.

- **Order is deliberate, not incidental.** `Evaluate` and `EvaluateAll` iterate the Builder's field names *sorted*, so the same URL always produces the same expression and `EvaluateAll` always reports the same missing field. Ranging the map directly reintroduces nondeterminism.

- **`Evaluate` vs. `EvaluateAll`.** `Evaluate` includes only the fields present in the URL; `EvaluateAll` requires *every* registered field to be present and non-empty, returning an error otherwise. That error's detail is the field NAME — never the `Field` itself, which holds filter funcs that `encoding/json` refuses to marshal.
