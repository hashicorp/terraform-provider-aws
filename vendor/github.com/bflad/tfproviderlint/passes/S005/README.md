# S005

The S005 analyzer reports cases of schemas which enables `Computed`
and configures `Default`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    Default:  /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}

# OR

&schema.Schema{
    Default: true,
}
```
