# S016

The S016 analyzer reports cases of schemas including `Set` without `TypeSet`,
which will fail schema validation.

## Flagged Code

```go
&schema.Schema{
    Set:  /* ... */,
    Type: schema.TypeList,
}
```

## Passing Code

```go
&schema.Schema{
    Set:  /* ... */,
    Type: schema.TypeSet,
}
```
