# S020

The S020 analyzer reports cases of schemas which enables only `Computed`
and enables `ForceNew`, which is invalid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    ForceNew: true,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```
