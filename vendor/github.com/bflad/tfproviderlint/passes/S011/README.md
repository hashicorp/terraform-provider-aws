# S011

The S011 analyzer reports cases of schemas which enables only `Computed`
and configures `DiffSuppressFunc`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed:         true,
    DiffSuppressFunc: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```
