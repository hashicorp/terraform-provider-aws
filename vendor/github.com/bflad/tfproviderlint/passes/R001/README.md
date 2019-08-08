# R001

The R001 analyzer reports a complex key argument for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set)
call. It is preferred to explicitly use a string literal as the key argument.

## Flagged Code

```go
keys := []string{"example1", "example2"}
values := []string{"value1", "value2"}

for idx, key := range keys {
    d.Set(key, values[idx])
}
```

## Passing Code

```go
d.Set("example1", "value1")
d.Set("example2", "value2")
```
