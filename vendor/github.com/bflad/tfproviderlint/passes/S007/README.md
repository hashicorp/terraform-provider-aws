# S007

The S007 analyzer reports cases of schemas which enables `Required`
and configures `ConflictsWith`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required:      true,
    ConflictsWith: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Required: true,
}

# OR

&schema.Schema{
    Optional:      true,
    ConflictsWith: /* ... */,
}
```
