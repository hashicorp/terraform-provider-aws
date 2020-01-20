# S004

The S004 analyzer reports cases of schemas which enables `Required`
and configures `Default`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required: true,
    Default:  /* ... */,
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
    Default:  /* ... */,
}
```
