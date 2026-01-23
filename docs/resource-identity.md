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

Some AWS resource types allow only a single instance in a given region,
or in a single account for global resource types.

Specify a Singleton Identity for a resource type by adding the annotation `@SingletonIdentity` to the resource type's declaration.

The Resource Identity attributes for Singleton Identities cannot be overridden.

### Parameterized Identity

Many AWS resource types can be uniquely identified by a combination of one or more attributes, such as a name, within their region (or account for global resource types).

Specify a Parameterized Identity using one or more `@IdentityAttribute("<attribute-name>")` annotations to the resource type's declaration.
The attributes `account_id` and `region` are always present in Parameterized Identities, so must not be specified in `@IdentityAttribute` annotations.

By default, the names of the Resource Identity attribute and the corresponding attribute on the resource match.
In some rare cases, the Resource Identity attribute cannot use the same name.
For example, in the resource type `aws_organizations_delegated_administrator`, one of the identifying resource attributes is `account_id`, which is reserved.
Instead, the Resource Identity attribute is `delegated_account_id`.
The annotation is `@IdentityAttribute("delegated_account_id", resourceAttributeName="account_id")`.

In some rare cases, an identifying attribute may allow `null` as a value.
For example, the resource type `aws_route53_record` has an optional attribute `set_identifier` that differentiates between individual records in a set, such as multiple addresses for the same domain name.
In simple routing cases, `set_identifier` is unused.
To specify this, add the annotation parameter `optional=true` to the `@IdentityAttribute` annotation.
As the `set_identifier` in `aws_route53_record` is optional,
the full annotation is `@IdentityAttribute("set_identifier", optional="true")`.

If a resource type has optional identifying attributes,
the default generated test should be the case where all optional values are `null`.
For other cases, manually create additional acceptance tests,
at a minimum the equivalent of the `Basic` test as well as the `ExistingResource` test when adding Resource Identity to an exisiting resource type.

If it is not possible to create a test case where all optional values are `null`,
add the annotation parameter `testNotNull=true` to the corresponding `@IdentityAttribute` annotation.

## Enabling Resource Identity on New Resource Type

New resource types with Resource Identity support are indicated by the annotation `@Testing(hasNoPreExistingResource=true)`.
This will exclude the [generation of tests](#acceptance-testing) related to adding Identity data to an existing resource.

## Adding Resource Identity to Existing Resource Type

When adding Resource Identity to an existing resource type, there are several annotations to add to the resource type declaration.

In order to [generate tests](#acceptance-testing) related to adding Identity data to an existing resource, we need to indicate the last version of the provider before Resource Identity was enabled on the resource type.
Add the annotation `@Testing(preIdentityVersion="<version>")`, where version is the last version of the provider **before** Resource Identity is added to the resource type.
For example, Resource Identity was added to `aws_batch_job_definition` in version 6.5.0, so the annotation is `preIdentityVersion="v6.4.0"`.

In some cases, even though a resource type has an ARN Identity or Singleton Identity, it also has an `id` attribute that is set to the same value.
Specify this by adding the annotation parameter `identityDuplicateAttributes="id"` to the identity annotation.
(This will always be the case for resource types implemented with the Plugin SDK, so setting `identityDuplicateAttributes="id"` is not necessary for those resource types.)
For example, the resource type `aws_rds_integration` has an `id` attribute that duplicates the `arn`.
The annotation is `@ArnIdentity(identityDuplicateAttributes="id")`.

In some rare cases, there will be multiple attributes that match the Identity attribute.
Specifcy this with the annotation parameter `identityDuplicateAttributes="<attr>[;<attr>]"`.
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
Specify this with the annotation `@Testing(importIgnore="...")` with a list of the attribute names separated by semi-colons (`;`).

### Manual Tests

In some rare cases, generating the Resource Identity acceptance tests cannot be done.
As much as possible, try to use the test setup created by the generator.
Add the annotation `@Testing(identityTest=false)` and add a comment above the annotation indicating why the generated acceptance tests could not be used.
If there are missing configuration or configuration variable annotations, consider creating a GitHub issue indicating what is missing.

If not using the generated tests, rename the test file to remove `gen` from the name to indicate that it is not generated.
