# R005

The R005 analyzer reports when multiple [`HasChange()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.HasChange) calls in a conditional can be combined into a single [`HasChanges()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.HasChanges) call.

## Flagged Code

```go
if d.HasChange("attr1") || d.HasChange("attr2") {
  // handle attr1 and attr2 changes
}
```

## Passing Code

```go
if d.HasChanges("attr1", "attr2") {
  // handle attr1 and attr2 changes
}
```
