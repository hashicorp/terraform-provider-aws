# Retry Package

The start of musings on a replacement for the Terraform Plugin SDK v2 `helper/retry` package.

### Example Usage

```go
for r := retry.Begin(); r.Continue(ctx); {
    if doSomething() {
        break
    }
}
```
