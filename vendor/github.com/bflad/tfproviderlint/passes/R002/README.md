# R002

The R002 analyzer reports likely extraneous uses of
star (`*`) dereferences for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) call. The `Set()` function automatically
handles pointers and `*` dereferences without `nil` checks can panic.

## Flagged Code

```go
var stringPtr *string

d.Set("example", *stringPtr)
```

## Passing Code

```go
var stringPtr *string

d.Set("example", stringPtr)
```
