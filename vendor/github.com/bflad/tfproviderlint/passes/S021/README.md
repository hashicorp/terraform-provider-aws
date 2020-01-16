# S021

The S021 analyzer reports cases of schemas including `ComputedWhen`, which should be removed.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    ComputedWhen: []string{"another_attr"},
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```
