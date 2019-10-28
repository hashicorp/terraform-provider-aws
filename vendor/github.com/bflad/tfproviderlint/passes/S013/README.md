# S013

The S013 analyzer reports cases of schemas which one of `Computed`,
`Optional`, or `Required` is not configured, which will fail provider
schema validation.

## Flagged Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Type: schema.TypeString,
    },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Computed: true,
        Type:     schema.TypeString,
    },
}

# OR

map[string]*schema.Schema{
    "attribute_name": {
        Optional: true,
        Type:     schema.TypeString,
    },
}

# OR

map[string]*schema.Schema{
    "attribute_name": {
        Required: true,
        Type:     schema.TypeString,
    },
}
```
