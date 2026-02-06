<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Generated Acceptance Tests

Acceptance tests for resource tagging and Resource Identity are generated based on a common pattern.

Test generation is enabled at the service level, in the service's `generate.go` file.
When test generation is enabled for a service, it can be disabled for a specific resource or data source type.

To enable resource tagging tests, add the following line:

```go
//go:generate go run ../../generate/tagstests/main.go
```

To disable resource tagging tests for a specific resource or data source type, add the following annotation to its source file:

```
@Testing(tagsTest=false)
```

To enable Resource Identity tests, add the following line:

```go
//go:generate go run ../../generate/identitytests/main.go
```

To disable Resource Identity tests for a specific resource type, add the following annotation to its source file:

```
@Testing(identityTest=false)
```

## Configuring Generated Tests

Generated acceptance tests require configuration via resource type annotations.
Some annotations apply to both resource tagging and Resource Identity tests,
while others apply to only one or the other.

### Referencing Functons and Variables

Some testing annotations allow referencing functions or variables in the current package or another package.
These use a common reference format that contains a Go package path and package alias in the following format:
`[<package path>;[<package alias>;]]<function name>`.

### Common Configurations

The following testing configurations apply to both resource tagging and Resource Identity tests.

#### PreCheck Functions

The existing tests for a resource type will have one or more functions in the `PreCheck` function.
All generated acceptance tests include the standard `acctest.PreCheck` PreCheck function.
In many cases, acceptance tests will require additional PreCheck functions.

If a test can only be run in specific regions, the PreCheck function `acctest.PreCheckRegion` can be used.
Specify this with the annotation `@Testing(preCheckRegion="<regions>)`,
where `<regions>` is a semi-colon separated list of one or more region names.

Most `PreCheck` functions have the signature `func(ctx context.Context, t *testing.T)`.
Specify them with the annotation `@Testing(preCheck=<reference>)`.
The reference optionally contains a Go package path and package alias, using the format
`[<package path>;[<package alias>;]]<function name>`.
Multiple `@Testing(preCheck)` annotations are allowed.

In some cases, the PreCheck function will have the signature `func(ctx context.Context, t *testing.T, region string)`.
Specify these with the annotation `Testing(preCheckWithRegion=<reference>)`.
The reference optionally contains a Go package path and package alias, using the format
`[<package path>;[<package alias>;]]<function name>`.
Multiple `@Testing(preCheckWithRegion)` annotations are allowed.

#### Required Environment Variables

If a test should be skipped unless an environment variable is set, but the value is not used in the test,
set the annotation `@Testing(requireEnvVar="<name>")`.
Multiple `@Testing(requireEnvVar)` annotations are allowed.

If a test should be skipped unless an environment variable is set, and the value is used in the test,
set the annotation `@Testing(requireEnvVarValue="<name>")`.
This will add a Terraform variable with the same name to the generated test configuration.
Multiple `@Testing(requireEnvVarValue)` annotations are allowed.

#### Alternate Initialized Providers

Some resource types require alternate providers configured either for an alternate AWS account or an alternate region.
If the test uses multiple regions, consider using the `region` attribute on resource in the alternate region instead of a separate provider instance.

To specify a provider instance configured for an alternate account, use the annotation `@Testing(useAlternateAccount=true)`.
This will add the `PreCheck` function `acctest.PreCheckAlternateAccount` as well as initializing a provider instance with the alias `awsalternate`.

To specify a provider instance configured for an alternate region, use the annotation `@Testing(altRegionProvider=true)`.
This will add the `PreCheck` function `acctest.PreCheckMultipleRegion` as well as initializing a provider instance with the alias `awsalternate`.

#### Exists and Destroy Checks

Most `Exists` functions used in acceptance tests take a pointer to the returned API object.
To specify the type of this parameter, use the annotation `@Testing(existsType=<reference>)`.
This references a Go type and package path with optional package alias, using the format
`<package path>;[<package alias>;]<function call>`.
For example, the S3 Object uses

```go
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3;s3.GetObjectOutput")
```

In some rare cases, there is no `Exists` function for a resource type.
To specify this, use the annotation `@Testing(hasExistsFunction=false)`.

Some older resource types use variants of the standard `Exists` and `DestroyCheck` functions that do not take a `testing.T` parameter.
In that case, add the annotations `@Testing(existsTakesT=false)` and `@Testing(destroyTakesT=false)`, respectively.

Some resource types use the no-op `CheckDestroy` function `acctest.CheckDestroyNoop`.
Use the annotation `@Testing(checkDestroyNoop=true)`.

#### Import State Test Steps

The generated acceptance tests use `ImportState` steps.
In most cases, these will work as-is.
To ignore the values of certain parameters when importing, set the annotation `@Testing(importIgnore="...")` to a list of the parameter names separated by semi-colons (`;`).

Some resource types do not support the Import operation.
To specify this, use the annotation `@NoImport`.

There are multiple methods for overriding the import ID, if needed.
To use the value of an existing variable, use the annotation `@Testing(importStateId=<var name>)`.
If the identifier can be retrieved from a specific resource attribute, use the annotation `@Testing(importStateIdAttribute=<attribute name>)`.
If the identifier can be retrieved from a `resource.ImportStateIdFunc`, use the annotation `@Testing(importStateIdFunc=<func name>)`.

#### Test Serialization

If the tests need to be serialized, use the annotation `@Testing(serialize=true)`.
If a delay is needed between serialized tests, also use the annotation `@Testing(serializeDelay=<duration>)` with a duration in the format used by [`time.ParseDuration()`](https://pkg.go.dev/time#ParseDuration).
For example, 3 minutes and 30 seconds is `3m30s`.

#### Terraform Variables

Most testing configurations take a single parameter, often a name or a domain name.
The most common case is parameter `rName` with a value generated by `sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)`, so this is the default.

If no `rName` is required, add the annotation `@Testing(generator=false)`.
If the test's Terraform configuration does not reference the generated `rName` variable, `tflint` automated checks will fail with a `terraform_unused_declarations` error.

Other values can be used by setting the `generator` to a reference to a function call.
The reference optionally contains a Go package path and package alias, using the format
`[<package path>;[<package alias>;]]<function call>`.
For example, the Service Catalog Portfolio uses a five-character long random string

```go
// @Testing(generator="github.com/hashicorp/terraform-plugin-testing/helper/acctest;sdkacctest;sdkacctest.RandString(5)")
```

Some acceptance tests also require a TLS key and certificate.
This can be included by setting the annotation `@Testing(tlsKey=true)`,
which will add the Terraform variables `certificate_pem` and `private_key_pem` to the configuration.
By default, the common name for the certificate is `example.com`.
To override the common name, set the annotation `@Testing(tlsKeyDomain=<reference>)` to reference an existing variable.
For example, the API Gateway v2 Domain Name sets the variable `rName` to `acctest.RandomSubdomain()`
and sets the annotation `@Testing(tlsKeyDomain=rName)` to reference it.

Some acceptance tests require a TLS ECDSA public key PEM.
This can be included by setting the annotation `@Testing(tlsEcdsaPublicKeyPem=true)`.
The Terraform variable name will be `rTlsEcdsaPublicKeyPem`.

Some acceptance tests related to networking require a random BGP ASN value.
This can be included by setting the annotation `@Testing(randomBsgAsn="<low end>;<high end>)`,
where `<low end>` and `<high end>` are the upper and lower bounds for the randomly-generated ASN value.
The Terraform variable name will be `rBgpAsn`.

Some acceptance tests related to networking require a random IPv4 address.
This can be included by setting the annotation `@Testing(randomIPv4Address="<CIDR range>)`.
The randomly-generated IPv4 address value will be contained within the `<CIDR range>`.
The Terraform variable name will be `rIPv4Address`.

No additional parameters can be defined currently.
If additional parameters are required, and cannot be derived from `rName`, the resource type must use manually created acceptance tests as described in the [Resource Tagging documentation](resource-tagging.md#manually-created-acceptance-tests).

### Resource Identity Test Configuration

The following annotations can be used to configure the generated Resource Identity tests.

#### Region Override Tests

By default, unless a resource type is global, Resource Identity will also be tested in the acceptance testing alternate region,
configured using the environment variable `AWS_ALTERNATE_REGION` and defaulting to `us-east-1`.
In some rare cases, for example if a service is only available in a specific region,
the test for the alternate region can be omitted by using the annotation `@Testing(identityRegionOverrideTest=false)`.

#### Duplicated `id` and Resource Identity Attributes

If the resource type is implemented using Plugin Framework;
has an ARN Identity, Singleton Identity, or Custom Inherent Region Identity;
and has one or more attributes that duplicates the ARN attribute,
add the `identityDuplicateAttributes="<attr>[;<attr>]"` parameter to the identity annotation.
For example, the resource type `aws_ssoadmin_application` has an ARN Identity and the `arn` attribute is duplicated by both `id` and `application_arn`, so the annotation is `@ArnIdentity(identityDuplicateAttributes="id;application_arn")`.

Otherwise, if the resource type has an `id` attribute that is set to the same value as an identity attribute,
add an `@Testing(idAttrDuplicates="<attribute_name>")` annotation.

#### Composed Attribute Values

Some resource attributes related to Resource Identity are composed from other attribute values.
This is typically the case for `arn` and in many cases `id`.

If the `arn` resource attribute can be exactly composed from known attribute values,
add the annotation `@ArnFormat(<format>)`, where `<format>` is the exact string to match, with the attribute values replaced by the attribute name surrounded by braces (`{` and `}`).
The Partition, Region, and Account ID are automatically added as needed.
For example, the ARN format for `aws_batch_job_definition` is `job-definition/{name}:{revision}`.

In the case that the resource type has an ARN value attribute, but does not have an ARN Identity,
add the parameter `attribute="<arn-attribute-name>"`.
For example, the resource type `aws_appflow_flow` uses the `name` attribute for the Resource Identity,
so the `@ArnFormat` annotation has the parameter `attribute="arn"`.

In the rare case that a resource type is not global, but the ARN value does not include the region,
add the parameter `global=true`.
For example, the resource type `aws_ssoadmin_application` has an ARN with out the region,
so the `@ArnFormat` annotation has the parameter `global=true`.

If the `id` resource attribute can be exactly composed from known attribute values,
add the annotation `@IdAttrFormat(<format>)`, where `<format>` is the exact string to match, with the attribute values replaced by the attribute name surrounded by braces (`{` and `}`).
For example, the ID format for `aws_iam_role_policy_attachment` is `{role}/{policy_arn}`.

#### Adding Resource Identity to an Existing Resource Type

When adding Resource Identity to an existing resource type, we require additional acceptance tests to ensure that the Resource Identity is added to the resource correctly.
This is specified by the annotation `@Testing(preIdentityVersion="<version>")`, where version is the last version of the provider **before** Resource Identity is added to the resource type.
For example, Resource Identity was added to `aws_batch_job_definition` in version 6.5.0, so the annotation is `preIdentityVersion="v6.4.0"`.

#### Enabling Resource Identity on a New Resource Type

When enabling Resource Identity on a new resource type,
add the annotation `@Testing(hasNoPreExistingResource=true)`.

#### Adding a New Resource Identity Schema Version

In some circumstances, a resource type requires an updated Resource Identity schema version.
This requires additional testing to ensure that the Resource Identity can be correctly updated to the new schema.
This is specified by adding one `@Testing(identityVersion="<identity-schema-version>;<provider-version>")` annotation per Resource Identity schema version.
Schema versions are zero-indexed.
Note that, unlike the `@Testing(preIdentityVersion)` annotation, `identityVersion` uses the **actual** provider version.
For example, Resource Identity was added to the resource type `aws_sqs_queue` in provider version 6.10.0 and the schema was updated in provider version 6.19.0. The annotations are:

```go
// @Testing(preIdentityVersion="v6.9.0")
// @Testing(identityVersion="0;v6.10.0")
// @Testing(identityVersion="1;v6.19.0")
```

#### Plannable Import Behavior

Using an `import` block in a Terraform configuration allows the Import action to be part of the Terraform plan.
In most cases, importing a resource should result in no change in the resource after the import.
However, there are some cases where the resulting plan will have a change to the resource.

If there are attributes that need to be ignored during import using the `@Testing(importIgnore)` annotation,
by default the plan is assumed to update the resource.
In some cases, for example if the attribute has a default value,
the planned import will not update the resource.
Add the annotation `@Testing(plannableImportAction="NoOp")`.

In some rare cases, the imported resource will result in a plan that replaces the resource.
This is often the case when the resource type contains secrets, such as a certificate or credential.
This situation should be avoided.

If you are adding Resource Identity to an existing resource type,
add the annotation `@Testing(plannableImportAction="Replace")`.
Consider opening an issue in the provider's GitHub repository reporting that the resource type causes recreation when importing.

If you are creating a new resource type, consider how the resource type can be support importing without recreating the resource.
If it is not possible, disable import support using the annotation `@NoImport`.

#### Error Cases

When Resource Identity was first introduced to the provider, there were some errors in the implementation.
The following annotations are documented for completeness and **should not be used in new implementations**.

The annotation `@V60SDKv2Fix`, `@Testing(v60NullValuesError)`, and `@Testing(v60RefreshError)` indicate that special handling and testing was needed to fix errors in the initial implementation.

### Resource Tagging Test Configuration

In some cases, the AWS tagging APIs for a service or specific resource have non-standard behavior.
The following annotations can work around these cases.

#### Empty String and Null Tag Values

Some services do not support tags with an empty string value.
In that case, use the annotation `@Testing(skipEmptyTags=true)`.

Some services do not support tags with a null string value.
In that case, use the annotation `@Testing(skipNullTags=true)`.

#### Tag Update Behavior

For some resource types, tags cannot be modified without recreating the resource.
Use the annotation `@Testing(tagsUpdateForceNew=true)`.

At least one resource type, the Service Catalog Provisioned Product, does not support removing tags.
This is likely an error on the AWS side.
Add the annotation `@Testing(noRemoveTags=true)` as a workaround.

Resource types which pass the result of `getTagsIn` directly onto their Update Input may have an error where ignored tags are not correctly excluded from the update.
Use the annotation `@Testing(tagsUpdateGetTagsIn=true)` if this causes an error.

#### Identifier Attributes

Some tests read the tag values directly from the AWS API.
If the resource type does not specify `identifierAttribute` in its `@Tags` annotation, specify a `@Testing(tagsIdentifierAttribute=<attribute name>)` annotation to identify which attribute value should be used by the `listTags` function.
If a resource type is also needed for the `listTags` function, also specify the `tagsResourceType` annotation.

## Terraform Configuration Templates for Tests

The generated acceptance tests use `ConfigDirectory` to specify the test configurations in a directory of Terraform `.tf` files.
The configuration files are generated from a [Go template](https://pkg.go.dev/text/template) file located in `testdata/tmpl/<name>_basic.gtpl`,
where `name` is the name of the resource type's implementation file without the `.go` extension.
For example, the ELB v2 Load Balancer's implementation file is `load_balancer.go`, so the template is `testdata/tmpl/load_balancer_basic.gtpl`.

To generate a configuration for a data source test, the generator reuses the configuration for the corresponding resource type.
Add an additional file `testdata/tmpl/<name>_data_source.gtpl` which contains only the data source block populated with the parameters needed to associate it with the resource.
For example, the ELB v2 Load Balancer's data source template is `testdata/tmpl/load_balancer_data_source.gtpl`.

For all resources and data sources declared in the configuration, unless the type is global, add the Go template directive `{{- template "region" }}` at the top of the resource declaration.

Replace the `tags` attribute with the Go template directive `{{- template "tags" . }}`.
When the configurations are generated, this will be replaced with the appropriate assignment to the `tags` attribute.
The `tags` attribute should be the last line of the resource or data source definition.

Tags should only be applied to the resource that is being tested.

For example:

```terraform
resource "aws_service_thing" "test" {
{{- template "region" }}

  name = var.rName

{{- template "tags" . }}
}
```

### Pre-Defined Configuration Sections

To aid in simplifying and standardizing Terraform configurations for testing, there are a number of pre-defined configuration sections that mirror the pre-defined config functions in the `acctest` package.
The configuration sections have the same name as the pre-defined config function.

To include the pre-defined configuration sections add a line like

```
{{ template <name> <parameters> }}
```

to the configuration template.
For example, to include the configuration section `acctest.ConfigVPCWithSubnets`, which takes the number of subnets as a parameter, add

```
{{ template "acctest.ConfigVPCWithSubnets" 2 }}
```

#### `acctest.ConfigVPCWithSubnets`

Takes the subnet count as a parameter.

Creates an `aws_vpc` resource with the name `test` as well as the requested number of `aws_subnet` resources with the name `test`.

#### `acctest.ConfigSubnets`

Takes the subnet count as a parameter.

Creates the requested number of `aws_subnet` resources with the name `test`.

#### `acctest.ConfigVPCWithSubnetsIPv6`

Takes the subnet count as a parameter.

Creates an `aws_vpc` resource with the name `test` as well as the requested number of `aws_subnet` resources with the name `test`.

#### `acctest.ConfigSubnetsIPv6`

Takes the subnet count as a parameter.

Creates the requested number of `aws_subnet` resources with the name `test`.

#### `acctest.ConfigAvailableAZsNoOptInDefaultExclude`

Defines an `aws_availability_zones` data source named `available` which lists availability zones in the current region which are available and not opt-in.

#### `acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI`

Defines an `aws_ami` data source named `amzn2-ami-minimal-hvm-ebs-x86_64` which returns AMI data for an Amazon Linux 2 instance with `x86_64` architecture.

#### `acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI`

Defines an `aws_ami` data source named `amzn2-ami-minimal-hvm-ebs-arm64` which returns AMI data for an Amazon Linux 2 instance with `arm64` architecture.

#### `acctest.configLatestAmazonLinux2HVMEBSAMI`

Prefer `acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI` or `acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI` instead.

Takes processor architecture as parameter.

Defines an `aws_ami` data source named `amzn2-ami-minimal-hvm-ebs-<architecture>` which returns AMI data for an Amazon Linux 2 instance with specified architecture.

#### `acctest.ConfigAlternateAccountProvider`

Use only in cases where a test configuration needs to reference an instance of the AWS Provider configured with an alternate account.

Defines a provider named `awsalternate`.
