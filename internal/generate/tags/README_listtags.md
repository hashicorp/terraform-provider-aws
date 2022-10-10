# listtags

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for listing resource tags. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `list_tags_gen.go` file

## Example Output

```go
// AmplifyListTags lists amplify service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func AmplifyListTags(conn *amplify.Amplify, identifier string) (KeyValueTags, error) {
    input := &amplify.ListTagsForResourceInput{
        ResourceArn: aws.String(identifier),
    }

    output, err := conn.ListTagsForResource(input)

    if err != nil {
        return New(nil), err
    }

    return AmplifyKeyValueTags(output.Tags), nil
}
```

## Implementing a New Generated Service

### Requirements

Before a new service can be added to the generator, the new service must:

- Have the `KeyValueTags` conversion functions implemented for the AWS Go SDK service type/map. See also the [`servicetags` generator README](../servicetags/README.md).
- Implement a function for listing resource tags (e.g. `ListTagsforResource`)
- Have the service included in `aws/internal/keyvaluetags/service_generation_customizations.go`, if not present the following compilation error will be seen:

```text
2019/09/03 09:22:21 error executing template: template: listtags:19:41: executing "listtags" at <ClientType>: error calling ClientType: unrecognized ServiceClientType: acmpca
```

Once the service has met all the requirements, in `main.go`:

- Add import for new service, e.g. `"github.com/aws/aws-sdk-go/service/athena"`
- Add service name to `serviceNames`, e.g. `athena`
- Run `go generate ./...` (or `make gen`) from the root of the repository to regenerate the code
- Run `go test ./...` (or `make test`) from the root of the repository to ensure the generated code compiles
- (Optional) Customize the service generation, if necessary (see below)

### Customizations

By default, the generator creates a `{SERVICE}ListTags()` function with the following structs and function calls:

- `{SERVICE}.ListTagsForResourceInput` struct with `ResourceArn` field for calling `ListTagsForResource()` API call

If these do not match the actual AWS Go SDK service implementation, the generated code will compile with errors. See the sections below for certain errors and how to handle them.

#### ServiceListTagsFunction

Given the following compilation error:

```text
./list_tags_gen.go:183:12: undefined: backup.ListTagsForResourceInput
./list_tags_gen.go:187:21: conn.ListTagsForResource undefined (type *backup.Backup has no field or method ListTagsForResource)
```

The function for listing resource tags must be updated. Add an entry within the `ServiceListTagsFunction()` function of the generator to customize the naming of the `ListTagsForResource()` function and matching `ListTagsForResourceInput` struct. In the above case:

```go
case "backup":
    return "ListTags"
```

#### ServiceListTagsInputIdentifierField

Given the following compilation error:

```text
./list_tags_gen.go:1118:3: unknown field 'ResourceArn' in struct literal of type transfer.ListTagsForResourceInput
```

The field name to identify the resource for tag listing must be updated. Add an entry within the `ServiceListTagsInputIdentifierField()` function of the generator to customize the naming of the `ResourceArn` field for the list tags input struct. In the above case:

```go
case "transfer":
    return "Arn"
```

#### ServiceListTagsOutputTagsField

Given the following compilation error:

```text
./list_tags_gen.go:206:38: output.Tags undefined (type *cloudhsmv2.ListTagsOutput has no field or method Tags)
```

The field name of the tags from the tag listing must be updated. Add an entry within the `ServiceListTagsOutputTagsField()` function of the generator to customize the naming of the `Tags` field for the list tags output struct. In the above case:

```go
case "cloudhsmv2":
    return "TagList"
```
