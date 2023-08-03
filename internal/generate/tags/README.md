# Tags Generator

This generator (`main.go`) generates files named `tags_gen`, such as `internal/service/ecs/tags_gen.go`. These files provide consistency in handling AWS resource tags. Initiate generating by calling `make gen` from the provider directory.

## Generator Directives

Control the code generated using flags of the directives that you include in a `generate.go` file for an individual service. For example, a file such as `internal/service/ecs/generate.go` may contain three directives (and a package declaration). This generator corresponds to the `../../generate/tags/main.go` directive. (The other directives are documented in their respective packages.)

```go
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeCapacityProviders
//go:generate go run ../../generate/tagresource/main.go
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ServiceTagsSlice -UpdateTags

package ecs
```

**NOTE:** A `generate.go` file should _only_ contain generator directives and a package declaration. Do not include related Go functions in this file.

## Generator Directive Flags

Some flags control generation a certain section of code, such as whether the generator generates a certain function. Other flags determine how generated code will work. Do not include flags where you want the generator to use the default value.

| Flag | Default | Description | Example Use |
| --- | --- | --- | --- |
| `GetTag` |  | Whether to generate GetTag | `-GetTag` |
| `ListTags` |  | Whether to generate ListTags | `-ListTags` |
| `ServiceTagsMap` |  | Whether to generate map service tags (use this or `ServiceTagsSlice`, not both) | `-ServiceTagsMap` |
| `ServiceTagsSlice` |  | Whether to generate slice service tags (use this or `ServiceTagsMap`, not both) | `-ServiceTagsSlice` |
| `UpdateTags` |  | Whether to generate UpdateTags | `-UpdateTags` |
| `ContextOnly` |  | Whether to generator only Context-aware functions | `-ContextOnly` |
| `ListTagsInFiltIDName` |  | List tags input filter identifier name | `-ListTagsInFiltIDName=resource-id` |
| `ListTagsInIDElem` | `ResourceArn` | List tags input identifier element | `-ListTagsInIDElem=ResourceARN` |
| `ListTagsInIDNeedSlice` |  | Whether list tags input identifier needs a slice | `-ListTagsInIDNeedSlice=yes` |
| `ListTagsOp` | `ListTagsForResource` | List tags operation | `-ListTagsOp=ListTags` |
| `ListTagsOutTagsElem` | `Tags` | List tags output tags element | `-ListTagsOutTagsElem=TagList` |
| `TagInCustomVal` |  | Tag input custom value | `-TagInCustomVal=aws.StringMap(updatedTags.IgnoreAWS().Map())` |
| `TagInIDElem` | `ResourceArn` | Tag input identifier element | `-TagInIDElem=ResourceARN` |
| `TagInIDNeedSlice` |  | Tag input identifier needs a slice | `-TagInIDNeedSlice=yes` |
| `TagInIDNeedValueSlice` |  | Tag input identifier needs a slice of values, rather than a slice of pointers | `-TagInIDNeedValueSlice=yes` |
| `TagInTagsElem` | Tags | Tag input tags element | `-TagInTagsElem=TagsList` |
| `TagKeyType` |  | Tag key type | `-TagKeyType=TagKeyOnly` |
| `TagOp` | `TagResource` | Tag operation | `-TagOp=AddTags` |
| `TagOpBatchSize` |  | Tag operation batch size | `-TagOpBatchSize=10` |
| `TagResTypeElem` |  | Tag resource type field | `-TagResTypeElem=ResourceType` |
| `TagType` | `Tag` | Tag type | `-TagType=TagRef` |
| `TagType2` |  | Second tag type | `-TagType2=TagDescription` |
| `TagTypeAddBoolElem` |  | Tag type additional boolean element | `-TagTypeAddBoolElem=PropagateAtLaunch` |
| `TagTypeIDElem` |  | Tag type identifier field | `-TagTypeIDElem=ResourceId` |
| `TagTypeKeyElem` | `Key` | Tag type key element | `-TagTypeKeyElem=TagKey` |
| `TagTypeValElem` | `Value` | Tag type value element | `-TagTypeValElem=TagValue` |
| `UntagInCustomVal` |  | Untag input custom value | `-UntagInCustomVal="&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAWS().Keys())}"` |
| `UntagInNeedTagKeyType` |  | Untag input needs tag key type | `-UntagInNeedTagKeyType=yes` |
| `UntagInNeedTagType` |  | Untag input needs tag type | `-UntagInNeedTagType` |
| `UntagInTagsElem` | `TagKeys` | Untag input tags element | `-UntagInTagsElem=Tags` |
| `UntagOp` | `UntagResource` | Untag operation | `-UntagOp=DeleteTags` |

## Legacy Documentation

(TODO: This needs to be updated...)

The `keyvaluetags` package is designed to provide a consistent interface for handling AWS resource key-value tags. Many of the AWS Go SDK services, implement their own Go struct with `Key` and `Value` fields (e.g. `athena.Tag`) while others simply use a map (e.g. `map[string]string`). These inconsistent implementations and numerous Go types makes the process of correctly working with each of the services a tedius, previously copy-paste-modify process.

This package instead implements a single `KeyValueTags` type, which covers all key-value handling logic such as merging tags and ignoring keys via functions on the single type. The underlying implementation is compatible with Go operations such as `len()`.

Full documentation for this package can be found on [GoDoc](https://godoc.org/github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags).

Many AWS Go SDK services that support tagging have their service-specific Go type conversion functions to and from `KeyValueTags` code generated. Converting from `KeyValueTags` to AWS Go SDK types is done via `{SERVICE}Tags()` functions on the type, while converting from AWS Go SDK types to the `KeyValueTags` type is done via `{SERVICE}KeyValueTags()` functions. For more information about this code generation, see the [`servicetags` README](README_servicetags.md).

Some AWS Go SDK services that have common tag listing functionality (such as `ListTagsForResource` API call), also have auto-generated list functions. For more information about this code generation, see the [`listtags` README](README_listtags.md).

Some AWS Go SDK services that have common tagging update functionality (such as `TagResource` and `UntagResource` API calls), also have auto-generated update functions. For more information about this code generation, see the [`updatetags` README](README_updatetags.md).

Any tagging functions that cannot be generated should be hand implemented in a service-specific source file (e.g. `iam_tags.go`) and follow the format of similar generated code wherever possible. The first line of the source file should be `// +build !generate`. This prevents the file's inclusion during the code generation phase.

## Code Structure

```text
aws/internal/keyvaluetags
├── generators
│   ├── createtags (generates create_tags_gen.go)
│   ├── gettag (generates get_tag_gen.go)
│   ├── listtags (generates list_tags_gen.go)
│   ├── servicetags (generates service_tags_gen.go)
│   └── updatetags (generates update_tags_gen.go)
├── key_value_tags_test.go (unit tests for core logic)
├── key_value_tags.go (core logic)
├── list_tags_gen.go (generated AWS Go SDK service list tag functions)
├── service_generation_customizations.go (shared AWS Go SDK service customizations for generators)
├── service_tags_gen.go (generated AWS Go SDK service conversion functions)
├── update_tags_gen.go (generated AWS Go SDK service tagging update functions)
└── <service name>_tags.go (any service-specific functions that cannot be generated)
```

<!-- markdownlint-disable -->
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

- Have the `KeyValueTags` conversion functions implemented for the AWS Go SDK service type/map. See also the [`servicetags` README](README_servicetags.md).
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

# updatetags

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for updating resource tags. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `update_tags_gen.go` file

## Example Output

```go
// AthenaUpdateTags updates athena service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func AthenaUpdateTags(conn *athena.Athena, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
    oldTags := New(oldTagsMap)
    newTags := New(newTagsMap)

    if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
        input := &athena.UntagResourceInput{
            ResourceARN: aws.String(identifier),
            TagKeys:     aws.StringSlice(removedTags.Keys()),
        }

        _, err := conn.UntagResource(input)

        if err != nil {
            return fmt.Errorf("untagging resource (%s): %s", identifier, err)
        }
    }

    if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
        input := &athena.TagResourceInput{
            ResourceARN: aws.String(identifier),
            Tags:        updatedTags.IgnoreAws().AthenaTags(),
        }

        _, err := conn.TagResource(input)

        if err != nil {
            return fmt.Errorf("tagging resource (%s): %s", identifier, err)
        }
    }

    return nil
}
```

## Implementing a New Generated Service

### Requirements

Before a new service can be added to the generator, the new service must:

- Have the `KeyValueTags` conversion functions implemented for the AWS Go SDK service type/map. See also the [`servicetags` README](README_servicetags.md).
- Implement a function for tagging (e.g. `TagResource`) and a function for untagging via keys (e.g. `UntagResource`)
- Have the service included in `aws/internal/keyvaluetags/service_generation_customizations.go`, if not present the following compilation error will be seen:

```text
2019/09/03 09:22:21 error executing template: template: listtags:19:41: executing "updatetags" at <ClientType>: error calling ClientType: unrecognized ServiceClientType: acmpca
```

Once the service has met all the requirements, in `main.go`:

- Add import for new service, e.g. `"github.com/aws/aws-sdk-go/service/athena"`
- Add service name to `serviceNames`, e.g. `athena`
- Add reflection handling to `ServiceClientType()` function, e.g.

```go
case "athena":
    funcType = reflect.TypeOf(athena.New)
```

- Run `go generate ./...` (or `make gen`) from the root of the repository to regenerate the code
- Run `go test ./...` (or `make test`) from the root of the repository to ensure the generated code compiles
- (Optional) Customize the service generation, if necessary (see below)

### Customizations

By default, the generator creates a `{SERVICE}UpdateTags()` function with the following structs and function calls:

- `{SERVICE}.TagResourceInput` struct with `ResourceArn` field and `Tags` field for calling `TagResource()` API call
- `{SERVICE}.UntagResourceInput` struct with `ResourceArn` field and `TagKeys` field for calling `UntagResource()` API call

If these do not match the actual AWS Go SDK service implementation, the generated code will compile with errors. See the sections below for certain errors and how to handle them.

#### ServiceTagFunction

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:704:13: undefined: datapipeline.TagResourceInput
aws/internal/keyvaluetags/update_tags_gen.go:709:17: conn.TagResource undefined (type *datapipeline.DataPipeline has no field or method TagResource)
```

The function for resource tagging must be updated. Add an entry within the `ServiceTagFunction()` function of the generator to customize the naming of the `TagResource()` function and matching `TagResourceInput` struct. In the above case:

```go
case "datapipeline":
    return "AddTags"
```

#### ServiceTagInputCustomValue

Given the following compilation errors:

```text
aws/internal/keyvaluetags/update_tags_gen.go:1994:4: cannot use updatedTags.IgnoreAws().KinesisTags() (type []*kinesis.Tag) as type map[string]*string in field value
```

or

```text
aws/internal/keyvaluetags/update_tags_gen.go:2534:4: cannot use updatedTags.IgnoreAws().PinpointTags() (type map[string]*string) as type *pinpoint.TagsModel in field value
```

The value of the tags for tagging must be transformed. Add an entry within the `ServiceTagInputCustomValue()` function of the generator to return the custom value. In the above case:

```go
case "kinesis":
	return "aws.StringMap(chunk.IgnoreAws().Map())"
```

#### ServiceTagInputIdentifierField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:296:4: unknown field 'ResourceArn' in struct literal of type athena.UntagResourceInput (but does have ResourceARN)
aws/internal/keyvaluetags/update_tags_gen.go:309:4: unknown field 'ResourceArn' in struct literal of type athena.TagResourceInput (but does have ResourceARN)
```

The field name to identify the resource for tagging must be updated. Add an entry within the `ServiceTagInputIdentifierField()` function of the generator to customize the naming of the `ResourceArn` field for the tagging and untagging input structs. In the above case:

```go
case "athena":
    return "ResourceARN"
```

#### ServiceTagInputIdentifierRequiresSlice

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:1296:4: cannot use aws.String(identifier) (type *string) as type []*string in field value
```

The value to identify the resource for tagging must be passed in a string slice. Add an entry within the `ServiceTagInputIdentifierRequiresSlice()` function of the generator to ensure that the value is passed as expected. In the above case

```go
case "ec2":
	return "yes"
```

#### ServiceTagInputTagsField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:382:4: unknown field 'Tags' in struct literal of type cloudhsmv2.TagResourceInput
```

The field name with the tags for tagging must be updated. Add an entry within the `ServiceTagInputTagsField()` function of the generator to customize the naming of the `Tags` field for the tagging input struct. In the above case:

```go
case "cloudhsmv2":
    return "TagList"
```

#### ServiceUntagFunction

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:691:13: undefined: datapipeline.UntagResourceInput
aws/internal/keyvaluetags/update_tags_gen.go:696:17: conn.UntagResource undefined (type *datapipeline.DataPipeline has no field or method UntagResource)
```

The function for resource untagging must be updated. Add an entry within the `ServiceUntagFunction()` function of the generator to customize the naming of the `UntagResource()` function and matching `UntagResourceInput` struct. In the above case:

```go
case "datapipeline":
    return "RemoveTags"
```

#### ServiceUntagInputTagsField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:369:4: unknown field 'TagKeys' in struct literal of type cloudhsmv2.UntagResourceInput
```

The field name with the tag keys for untagging must be updated. Add an entry within the `ServiceUntagInputTagsField()` function of the generator to customize the naming of the `TagKeys` field for the untagging input struct. In the above case:

```go
case "cloudhsmv2":
    return "TagKeyList"
```

#### ServiceUntagInputCustomValue

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:523:4: cannot use updatedTags.IgnoreAws().CloudfrontTags() (type []*cloudfront.Tag) as type *cloudfront.Tags in field value
```

The value of the tags for untagging must be transformed. Add an entry within the `ServiceUntagInputCustomValue()` function of the generator to return the custom value. In the above case:

```go
case "cloudfront":
	return "&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAws().Keys())}"
```
