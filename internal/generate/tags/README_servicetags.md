# servicetags

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for converting service tag/map types to/from `KeyValueTags`. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `service_tags_gen.go` file

## Example Output

For services with a specific Go type:

```go
// AthenaTags returns athena service tags.
func (tags KeyValueTags) AthenaTags() []*athena.Tag {
    result := make([]*athena.Tag, 0, len(tags))

    for k, v := range tags.Map() {
        tag := &athena.Tag{
            Key:   aws.String(k),
            Value: aws.String(v),
        }

        result = append(result, tag)
    }

    return result
}

// AthenaKeyValueTags creates KeyValueTags from athena service tags.
func AthenaKeyValueTags(tags []*athena.Tag) KeyValueTags {
    m := make(map[string]*string, len(tags))

    for _, tag := range tags {
        m[aws.StringValue(tag.Key)] = tag.Value
    }

    return New(m)
}
```

For services that implement a map instead:

```go
// AmplifyTags returns amplify service tags.
func (tags KeyValueTags) AmplifyTags() map[string]*string {
    return aws.StringMap(tags.Map())
}

// AmplifyKeyValueTags creates KeyValueTags from amplify service tags.
func AmplifyKeyValueTags(tags map[string]*string) KeyValueTags {
    return New(tags)
}
```

## Implementing a New Generated Service

- In `main.go`: Add service name, e.g. `athena`, to one of the implementation handlers
    - Use `sliceServiceNames` if the AWS Go SDK service implements a specific Go type such as `Tag`
    - Use `mapServiceNames` if the AWS Go SDK service implements `map[string]*string`
- Run `go generate ./...` (or `make gen`) from the root of the repository to regenerate the code
- Run `go test ./...` (or `make test`) from the root of the repository to ensure the generated code compiles
- (Optional, only for services with a specific Go type such as `Tag`) Customize the service generation, if necessary (see below)

### Customizations

By default, the generator creates `KeyValueTags.{SERVICE}Tags()` and `{SERVICE}KeyValueTags()` functions with the following expectations:

- `Tag` struct with `Key` field and `Value` field

If these do not match the actual AWS Go SDK service implementation, the generated code will compile with errors. See the sections below for certain errors and how to handle them.

#### ServiceTagType

Given the following compilation error:

```text
aws/internal/keyvaluetags/service_tags_gen.go:397:43: undefined: appmesh.Tag
aws/internal/keyvaluetags/service_tags_gen.go:398:20: undefined: appmesh.Tag
aws/internal/keyvaluetags/service_tags_gen.go:401:11: undefined: appmesh.Tag
aws/internal/keyvaluetags/service_tags_gen.go:413:34: undefined: appmesh.Tag
```

The Go type that represents a tag must be updated. Add an entry within the `ServiceTagType()` function of the generator to customize the naming of the Go type. In the above case:

```go
case "appmesh":
    return "TagRef"
```

#### ServiceTagTypeKeyField

Given the following compilation error:

```text
aws/internal/keyvaluetags/service_tags_gen.go:1563:4: unknown field 'Key' in struct literal of type kms.Tag
aws/internal/keyvaluetags/service_tags_gen.go:1578:24: tag.Key undefined (type *kms.Tag has no field or method Key)
```

The field name to identify the tag key within the Go type for tagging must be updated. Add an entry within the `ServiceTagTypeKeyField` function of the generator to customize the naming of the `Key` field for the tagging Go type. In the above case:

```go
case "kms":
    return "TagKey"
```

#### ServiceTagTypeValueField

Given the following compilation error:

```text
aws/internal/keyvaluetags/service_tags_gen.go:1564:4: unknown field 'Value' in struct literal of type kms.Tag
aws/internal/keyvaluetags/service_tags_gen.go:1578:39: tag.Value undefined (type *kms.Tag has no field or method Value)
```

The field name to identify the tag value within the Go type for tagging must be updated. Add an entry within the `ServiceTagTypeValueField` function of the generator to customize the naming of the `Value` field for the tagging Go type. In the above case:

```go
case "kms":
    return "TagValue"
```
