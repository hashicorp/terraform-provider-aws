# V005

_This terraform-plugin-sdk (v1) analyzer has been removed in tfproviderlint v0.30.0._

The V005 analyzer reports usage of the deprecated [ValidateJsonString](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#ValidateJsonString) validation function that should be replaced with [StringIsJSON](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringIsJSON).

## Flagged Code

```go
ValidateFunc: validation.ValidateJsonString,
```

## Passing Code

```go
ValidateFunc: validation.StringIsJSON,
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V005` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V005
ValidateFunc: validation.ValidateJsonString,
```
