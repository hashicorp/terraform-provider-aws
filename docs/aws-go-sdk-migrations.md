# AWS SDK For Go Migration Guide

AWS has [announced](https://aws.amazon.com/blogs/developer/announcing-end-of-support-for-aws-sdk-for-go-v1-on-july-31-2025/) that V1 of the SDK for Go will enter maintenance mode on July 31, 2024 and reach end-of-life on July 31, 2025.
While the AWS Terraform provider already [requires all net-new services](./aws-go-sdk-versions.md) to use AWS SDK V2 service clients, there remains a substantial number of existing services utilizing V1.

Over time maintainers will be migrating impacted services to adopt AWS SDK for Go V2.
For community members interested in contributing to this effort, this guide documents the common patterns required to migrate a service.

!!! tip
    The list of remaining services which require migration can be found in [this meta issue](https://github.com/hashicorp/terraform-provider-aws/issues/32976).

## Pre-Requisites

### Re-generate Service Client

When fully replacing the client, [`names/data/names_data.hcl`](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/data/names_data.hcl) should be updated to remove the v1 indicator and add v2 (ie. delete the `1` in the `ClientSDKV1` column and add a `2` in the `ClientSDKV2` column).
Once complete, re-generate the client.

```console
go generate ./internal/conns/...
```

### Add an `EndpointID` Constant

When a service is first migrated, a `{ServiceName}EndpointID` constant must be added to [`names/names.go`](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/names.go) manually.
Be sure to preserve alphabetical order.

The AWS SDK for Go V1 previously exposed these as constants, but V2 does not.
This constant is used in common acceptance testing pre-checks, such as `acctest.PreCheckPartitionHasService`.

### Patch Generation

The `awssdkpatch` tool can be used to automate parts of the AWS SDK migration.
Applying the generated patch will likely leave the provider in a state which does not compile.
As such, __this step is optional__ but can significantly reduce the amount of time spent on the steps below by applying changes without any manual intervention.

To apply a patch use the `awssdkpatch-apply` target, with the service to be migrated set to the `PKG` variable:

```console
PKG=ec2 make awssdkpatch-apply
```

You may also optionally generate the patch and use [`gopatch`](https://github.com/uber-go/gopatch) to preview differences before modfiying any files.

```console
make awssdkpatch-gen PKG=ec2
gopatch -d -p awssdk.patch ./internal/service/ec2/...
```

#### Custom options

To set additional `awssdkpatch` flags during patch generation, use the `AWSSDKPATCH_OPTS` environment variable.

```console
make awssdkpatch-gen PKG=ec2 AWSSDKPATCH_OPTS="-multiclient"
```

## Imports

In each go source file with a V1 SDK import, the library should be replaced with V2:

```go
// Remove
github.com/aws-sdk-go/service/<service>
```

```go
// Add
github.com/aws-sdk-go-v2/service/<service>
awstypes github.com/aws-sdk-go-v2/service/<service>/types
```

If the `aws` or `arn` packages are used, these should also be upgraded.

```
// Remove
github.com/aws-sdk-go/aws
github.com/aws-sdk-go/aws/arn
```

```
// Add
github.com/aws-sdk-go-v2/aws
github.com/aws-sdk-go-v2/aws/arn
```

## Client

Once the generated client is updated the following adjustments can be made.

### Initialization From `meta`

This is typically one of the first lines in each CRUD method.

```go
// Remove
conn := meta.(*conns.AWSClient).<service>Conn(ctx) 
```

```go
// Add
conn := meta.(*conns.AWSClient).<service>Client(ctx) 
```

### Passing Client References

Once initialized, the client may be passed into other functions, such as finders.

```go
// Remove
<service>.<Service>
```

```go
// Add
<service>.Client
// ie. ssoadmin.SSOAdmin becomes ssoadmin.Client
```

### Generated Tagging Functions

The generated tagging functions should be updated to use the new client by passing `-AWSSDKVersion=2`.
The following line should be updated in `generate.go`:

```go
//go:generate go run ../../generate/tags/main.go <existing flags> -AWSSDKVersion=2
```

Once updated, re-generate with:

```console
go generate ./internal/service/<service>/...
```

### `WithContext` Methods

All operations using the `WithContext` methods can be replaced, as the suffix is removed in V2 and context is passed everywhere by default.

```go
// Remove
conn.<Action>WithContext
```

```go
// Add
conn.<Action
// ie. conn.CreateApplicationWithContext becomes conn.CreateApplication
```

## Errors

Typically error types have moved into the services `types` package.
There are multiple variants which can be migrated.

```go
// Remove
if tfawserr.ErrCodeEquals(err, <service>.ErrCode<error>) {
```

```go
// Add
if errs.IsA[*awstypes.<error>](err) {
```

For example,

```go
if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
```

becomes:

```go
if errs.IsA[*awstypes.ResourceNotFoundException](err) {
```

The `ErrCodeContains` equivalent can also be migrated.

## AWS Helpers

These are simple 1-to-1 renames.

```go
// Remove
aws.StringValue
```

```go
// Add
aws.ToString
```

Similar variants for `Int32`, `Int64`, `Float32`, `Float64`, and `Bool` should also be changed.

## AWS Types

In general, the types for objects, enums, and errors move from the service package itself into the nested `types` package for each service.
This requires replacing the package prefix in several places.
As a starting point, any type reference which isn’t an Input/Output struct can likely be switched.

```go
// Remove
<service>.StatusInProgress
```

```go
// Add
awstypes.StatusInProgress
```

Individual services vary in which types are moved into the dedicated subpackage, so a programmatic replacement of all occurrences may require some manual adjustment afterward.

### Enum Types

Enums are now represented with custom types, rather than as strings.
Because of this type change it may be necessary to convert from or to string values depending on the direction in which data is being exchanged.

#### Input Structs

Input structs will now require the value to be wrapped in the enum type.

```go
// Remove
input.Thing = aws.String(v.(string))
```

```go
// Add
input.Thing = awstypes.Thing(v.(string))
```

#### Acceptance Test Attribute Checks

Acceptance tests where enums are passed to configuration functions or used for attribute checks will need to be converted from the enum type back into a string.

```go
// Remove
resource.TestCheckResourceAttr(resourceName, "thing", <service>.Thing),
```

```go
// Add
resource.TestCheckResourceAttr(resourceName, "thing", string(awstypes.Thing)),
```

#### Validation Functions

Validation functions which previously used the `StringInSlice` helper can now use a generic equivalent that works with custom string types.

```go
// Remove
ValidateFunc: validation.StringInSlice(<service>.Thing_Values(), false),
```

```go
// Add
ValidateDiagFunc: enum.Validate[awstypes.Thing](),
```

## Acceptance Testing `PreCheckPartitionHasService`

With V1, this check relies on the endpoint ID constant included in the SDK.
These are not included in the V2 SDK, but can be replaced with a constant from the `names` package.

```go
// Remove
acctest.PreCheckPartitionHasService(t, <service>.EndpointsID),
```

```go
// Add
acctest.PreCheckPartitionHasService(t, names.<Service>EndpointID),
```

For example,

```
acctest.PreCheckPartitionHasService(t, ssoadmin.EndpointsID),
```

becomes:

```go
acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID),
```

## Pagination

V2 of the AWS SDK introduces new [paginator](https://aws.github.io/aws-sdk-go-v2/docs/making-requests/#using-paginators) helpers.
Any `List*Pages` methods called with V1 of the SDK will need to replace the pagination function argument with the syntax documented in the AWS SDK for Go V2 documentation.

## Waiters

V2 of the AWS SDK introduces new [waiter](https://aws.github.io/aws-sdk-go-v2/docs/making-requests/#using-waiters) helpers.
Any `Wait*` methods called with V1 of the SDK will need to be replaced with the syntax documented in the AWS SDK for Go V2 documentation.

## Sweepers

All [sweepers](./running-and-writing-acceptance-tests.md#sweeper-checklists) should be updated to use the V2 SDK.
This should be similar to updating the resources themselves with steps like [updating imports](#imports), swapping the [service client](#client), adjusting [error handling](#errors), and adopting the appropriate [AWS `types` subpackage](#aws-types).
