# R003

The R003 analyzer reports likely extraneous uses of `Exists`
functions for a resource. `Exists` logic can be handled inside the `Read` function
to prevent logic duplication.

## Flagged Code

```go
func resourceExampleThingExists(d *schema.ResourceData, meta interface{}) (bool, error) { /* ... */ }

func resourceExampleThingRead(d *schema.ResourceData, meta interface{}) error { /* ... */ }

&schema.Resource{
    Exists: resourceExampleThingExists,
    Read:   resourceExampleThingRead,
    /* ... */
}
```

## Passing Code

```go
func resourceExampleThingRead(d *schema.ResourceData, meta interface{}) error { /* ... */ }

&schema.Resource{
    Read: resourceExampleThingRead,
    /* ... */
}
```
