# S002

The S002 analyzer reports cases of schemas which enables both `Required`
and `Optional`, which will fail provider schema validation.

## Flagged Code

```go
_ = schema.Schema{
    Required: true,
    Optional: true,
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
}
```
