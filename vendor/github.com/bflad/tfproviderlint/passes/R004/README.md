# R004

The R004 analyzer reports incorrect types for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) call value.
The `Set()` function only supports a subset of basic types, slices and maps of that
subset of basic types, and the [`schema.Set`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#Set) type.

## Flagged Code

```go
var t time.Time

d.Set("example", t)
```

## Passing Code

```go
var t time.Time

d.Set("example", t.Format(time.RFC3339))
```
