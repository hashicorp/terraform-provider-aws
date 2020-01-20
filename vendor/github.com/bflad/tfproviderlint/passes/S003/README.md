# S003

The S003 analyzer reports cases of schemas which enables both `Required`
and `Computed`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required: true,
    Computed: true,
}
```

## Passing Code

```go
&schema.Schema{
    Required: true,
}

# OR

&schema.Schema{
    Optional: true,
    Computed: true,
}
```
