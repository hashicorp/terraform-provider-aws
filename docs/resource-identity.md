<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Enabling Resource Identity for a Resource Type

Terraform version 1.12 introduced the concept of Resource Identity.
This is structured data which can uniquely identify a resource.

## Types of Resource Identity

There are several categories of Resource Identity in the AWS Provider, depending on how remote resources are identified in the appropriate AWS API.

### ARN Identity

Many AWS resource types support Amazon Resource Names (ARNs) that can uniquely identify a remote resource within AWS.
However, a resource type should only use an ARN Identity if the AWS APIs for this resource type take the ARN as the parameter to identify the remote resource.
Otherwise, the Resource Identity should use a [Parameterized Identity](#parameterized-identity).

Specify an ARN Identity for a resource type by adding the annotation `@ArnIdentity` to the resource type's declaration.

By default, the resource attribute and the Resource Identity attribute are named `arn`.
To override this, add the name of the attribute to the `@ArnIdentity` annotation.
For example, the resource type `aws_acmpca_policy` uses the attribute `resource_arn`,
so the annotation is `@ArnIdentity("resource_arn")`.

### Singleton Identity

TODO

### Parameterized Identity

TODO

## Enabling Resource Identity on New Resource Type

New resource types with Resource Identity support are indicated by the annotation `@Testing(hasNoPreExistingResource=true)`.
This will exclude the [generation of tests](#acceptance-testing) related to adding Identity data to an existing resource.

## Adding Resource Identity to Existing Resource Type

When adding Resource Identity to an existing resource type, there are several annotations to add to the resource type declaration.

In order to [generate tests](#acceptance-testing) related to adding Identity data to an existing resource, we need to indicate the last version of the provider before Resource Identity was enabled on the resource type.
Add the annotation `@Testing(preIdentityVersion="<version>")`, where version is the last version of the provider **before** Resource Identity is added to the resource type.
For example, Resource Identity was added to `aws_batch_job_definition` in version 6.5.0, so the annotation is `preIdentityVersion="v6.4.0"`.

In some cases, even though a resource type has an ARN Identity, it also has an `id` attribute that is set to the ARN value.
Specifcy this with the annotation `@ArnIdentity(identityDuplicateAttributes="id")`
(This will always be the case for resource types implemented with the Plugin SDK, so setting `identityDuplicateAttributes="id"` is not necessary for those resource types.)

In some rare cases, there will be multiple attributes that match the Identity attribute.
Specifcy this with the annotation `@ArnIdentity(identityDuplicateAttributes="<attr>[;<attr>]")`.
For example, the resource type `aws_ssoadmin_application` has both an `id` attribute and a deprecated alternate ARN attribute `application_arn`.
The annotation is `@ArnIdentity(identityDuplicateAttributes="id;application_arn")`.

## Updating Resource Identity Schema

TODO

## Acceptance Testing

Acceptance tests for Resource Identity are generated, as they follow the same pattern but have some subtleties.
See the [acceptance test generation documentation](acc-test-generation.md) for more information on generating Resource Identity tests.
Some common annotations are documented below.

By convention, the `Exists` check function returns a value from an API call for the remote resource.
This type is specified using the annotation `@Testing(existsType=<reference>)`.
This references a Go type and package path with optional package alias, using the format
`<package path>;[<package alias>;]<function call>`.
For example, the S3 Object uses

```go
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3;s3.GetObjectOutput")
```

Many acceptance tests make use of `PreCheck` functions.
The default `acctest.PreCheck` function is always included.
To specify other `PreCheck` functions, specify either
`@Testing(preCheck=<reference>)` if the function has the signature `func(ctx context.Context, t *testing.T)`, or
`Testing(preCheckWithRegion=<reference>)` if the function has the signature `func(ctx context.Context, t *testing.T, region string)`.

Some acceptance tests must ignore attribute differences when importing a resource.
Specify this with the annotation `@Testing(importIgnore="...")` with a list of the atribute names separated by semi-colons (`;`).

### Manual Tests

In some rare cases, generating the Resource Identity acceptance tests cannot be done.
As much as possible, try to use the test setup created by the generator.
Add the annotation `@Testing(identityTest=false)` and add a comment above the annotation indicating why the generated acceptance tests could not be used.
If there are missing configuration or configuration variable annotations, consider creating a GitHub issue indicating what is missing.

If not using the generated tests, rename the test file to remove `gen` from the name to indicate that it is not generated.
