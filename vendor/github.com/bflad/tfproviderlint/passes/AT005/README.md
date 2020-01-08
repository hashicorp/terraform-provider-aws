# AT005

The AT005 analyzer reports test function names (`Test` prefix) that contain
`resource.Test()` or `resource.ParallelTest()`, which should be named with
the TestAcc prefix.

## Flagged Code

```go
func TestExampleThing_basic(t *testing.T) {
  resource.Test(/* ... */)
}

func TestExampleWidget_basic(t *testing.T) {
  resource.ParallelTest(/* ... */)
}
```

## Passing Code

```go
func TestAccExampleThing_basic(t *testing.T) {
  resource.Test(/* ... */)
}

func TestAccExampleWidget_basic(t *testing.T) {
  resource.ParallelTest(/* ... */)
}
```
