# S008

The S008 analyzer reports cases of `TypeList` or `TypeSet` schemas configuring `Default`,
which will fail schema validation.

## Flagged Code

```go
&schema.Schema{
    Type:    schema.TypeList,
    Default: /* ... */,
}

&schema.Schema{
    Type:    schema.TypeSet,
    Default: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Type: schema.TypeList,
}

&schema.Schema{
    Type: schema.TypeSet,
}
```
