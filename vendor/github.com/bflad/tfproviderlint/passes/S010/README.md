# S010

The S010 analyzer reports cases of schemas which enables only `Computed`
and configures `ValidateFunc`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    ValidateFunc: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```
