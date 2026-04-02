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
at a minimum the equivalent of the `Basic` test as well as the `ExistingResource` test when adding Resource Identity to an existing resource type.

If it is not possible to create a test case where all optional values are `null`,
add the annotation parameter `testNotNull=true` to the corresponding `@IdentityAttribute` annotation.

#### What **Not** to Include in Resource Identity

The attributes used in Resource Identity should _only_ be those used to uniquely identify a resource.
In some cases, when adding Resource Identity to an existing resource type, the existing `id` attribute or the identifier used for importing a resource may include one or more values intended to preserve write-only values when refreshing or importing a resource.
These should not be included in Resource Identity.

For example, the resource type `aws_s3_bucket_acl` has an attribute `acl` which represents the name of a predefined permissions grant.
An S3 Bucket can have only one Bucket ACL, and the predefined grant name does not identify a Bucket ACL.
While the `acl` value is part of the `id` attribute, it should not be part of the Resource Identity.

#### Multiple Identity Attributes

In order to support importing resources by ID from the command line or in `import` blocks and work with `id` attributes composed from multiple fields,
Resource Identities with multiple attributes need a handler struct to perform the mapping.

##### Plugin Framework

In order to parse an import ID, Framework-based resource types must define a struct which implements the interface `inttypes.ImportIDParser`.

```go
type ImportIDParser interface {
	Parse(id string) (string, map[string]string, error)
}
```

The function `Parse` takes the import ID as a parameter and returns:

1. The value to be assigned to the `id` attribute, if any (see below)
1. A `map[string]string` of the resource attributes to be set
1. Any error

The name of this struct is set in the annotation `@ImportIDHandler("<struct name>")`.

For example, the resource type `aws_vpc_security_group_vpc_association` has an import ID handler as follows:

```go
var _ inttypes.ImportIDParser = securityGroupVPCAssociationImportID{}

type securityGroupVPCAssociationImportID struct{}

func (securityGroupVPCAssociationImportID) Parse(id string) (string, map[string]string, error) {
	sgID, vpcID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <security-group-id>"+intflex.ResourceIdSeparator+"<vpc-id>", id)
	}

	result := map[string]string{
		"security_group_id": sgID,
		names.AttrVPCID:     vpcID,
	}

	return id, result, nil
}
```

and has the annotation `@ImportIDHandler("securityGroupVPCAssociationImportID")`.

In some cases, the resource import will also need to set an `id` attribute composed from multiple fields.
In this case, the import ID parser must also implement the interface `inttypes.FrameworkImportIDCreator`.

```go
type FrameworkImportIDCreator interface {
	Create(ctx context.Context, state tfsdk.State) string
}
```

The function `Create` takes the state values and returns a single string value.

This is specified by the annotation parameter `setIDAttribute=true` on the `@ImportIDHandler` annotation.

For example, the resource type `aws_cloudfrontkeyvaluestore_key`has an import ID handler equivalent to:

```go
var (
	_ inttypes.ImportIDParser           = securityGroupVPCAssociationImportID{}
	_ inttypes.FrameworkImportIDCreator = securityGroupVPCAssociationImportID{}
)

type securityGroupVPCAssociationImportID struct{}

func (securityGroupVPCAssociationImportID) Parse(id string) (string, map[string]string, error) {
	kvsARN, key, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <key-value-store-arn>"+intflex.ResourceIdSeparator+"<key>", id)
	}

	result := map[string]string{
		"key_value_store_arn": kvsARN,
		names.AttrKey:         key,
	}

	return id, result, nil
}

func (securityGroupVPCAssociationImportID) Create(ctx context.Context, state tfsdk.State) string {
	parts := make([]string, 0, keyResourceIDPartCount)

	var attrVal types.String

	state.GetAttribute(ctx, path.Root("key_value_store_arn"), &attrVal)
	parts = append(parts, attrVal.ValueString())

	state.GetAttribute(ctx, path.Root(names.AttrKey), &attrVal)
	parts = append(parts, attrVal.ValueString())

	return strings.Join(parts, intflex.ResourceIdSeparator)
}
```

and has the annotation `@ImportIDHandler("securityGroupVPCAssociationImportID", setIDAttribute=true)`.

##### Plugin SDK

For resource types implemented using the Plugin SDK, the import ID handler must both parse the import ID and create the `id` attribute composed from multiple fields.
The import ID handler must implement the interface `inttypes.SDKv2ImportID`

```go
type SDKv2ImportID interface {
	Parse(id string) (string, map[string]string, error)
	Create(d *schema.ResourceData) string
}
```

The function `Parse` takes the import ID as a parameter and returns:

1. The value to be assigned to the `id` attribute
1. A `map[string]string` of the resource attributes to be set
1. Any error

The function `Create` takes the state values and returns a single string value.

The name of this struct is set in the annotation `@ImportIDHandler("<struct name>")`.

For example, a number of resource types associated with S3 Buckets, such as `aws_s3_bucket_logging` and `aws_s3_bucket_versioning` share an import ID handler equivalent to:

```go
var _ inttypes.SDKv2ImportID = resourceImportID{}

type resourceImportID struct{}

func (resourceImportID) Create(d *schema.ResourceData) string {
	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	if expectedBucketOwner == "" {
		return bucket
	}

	parts := []string{bucket, expectedBucketOwner}
	return strings.Join(parts, resourceIDSeparator)
}

func (resourceImportID) Parse(id string) (string, map[string]string, error) {
	bucket, expectedBucketOwner, err := parseResourceID(id)
	if err != nil {
		return id, nil, err
	}

	results := map[string]string{
		names.AttrBucket: bucket,
	}
	if expectedBucketOwner != "" {
		results[names.AttrExpectedBucketOwner] = expectedBucketOwner
	}

	return id, results, nil
}

func parseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, resourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected BUCKET or BUCKET%[2]sEXPECTED_BUCKET_OWNER", id, resourceIDSeparator)
}
```

## Enabling Resource Identity on New Resource Type

New resource types with Resource Identity support are indicated by the annotation `@Testing(hasNoPreExistingResource=true)`.
This will exclude the [generation of tests](#acceptance-testing) related to adding Identity data to an existing resource.

Using the [`skaff`](skaff.md) provider scaffolding tool, as recommended, to create a new resource type will add suggested annotations.

## Adding Resource Identity to Existing Resource Type

When adding Resource Identity to an existing resource type, there are several annotations to add to the resource type declaration.
See the [Adding Resource Identity Support Guide](add-resource-identity-support.md) for steps to take.

In order to [generate tests](#acceptance-testing) related to adding Identity data to an existing resource, we need to indicate the last version of the provider before Resource Identity was enabled on the resource type.
Add the annotation `@Testing(preIdentityVersion="<version>")`, where version is the last version of the provider **before** Resource Identity is added to the resource type.
For example, Resource Identity was added to `aws_batch_job_definition` in version 6.5.0, so the annotation is `preIdentityVersion="v6.4.0"`.

In some cases, even though a resource type has an ARN Identity, a Single-Attribute Parameterized Identity, or a Singleton Identity, it also has an `id` attribute that is set to the same value.
Specify this by adding the annotation parameter `identityDuplicateAttributes="id"` to the identity annotation.
(This will always be the case for resource types implemented with the Plugin SDK, so setting `identityDuplicateAttributes="id"` is not necessary for those resource types.)
For example, the resource type `aws_rds_integration` has an `id` attribute that duplicates the `arn`.
The annotation is `@ArnIdentity(identityDuplicateAttributes="id")`.

In some rare cases, there will be multiple attributes that match the Identity attribute.
Specify this with the annotation parameter `identityDuplicateAttributes="<attr>[;<attr>]"`.
For example, the resource type `aws_ssoadmin_application` has both an `id` attribute and a deprecated alternate ARN attribute `application_arn`.
The annotation is `@ArnIdentity(identityDuplicateAttributes="id;application_arn")`.

## Acceptance Testing

Acceptance tests for Resource Identity are generated, as they follow the same pattern but have some subtleties.
See the [acceptance test generation documentation](acc-test-generation.md) for more information on generating Resource Identity tests.
Some common annotations are documented below.

By convention, the `Exists` check function returns a value from an API call for the remote resource.
This type is specified using the annotation `@Testing(existsType=<reference>)`.
This references a Go type and package path with optional package alias, using the format
`<package path>;[<package alias>;]<type>`.
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

## Updating Resource Identity Schema

In some rare cases, it will be necessary to change the Resource Identity schema.
This may be due to changes in the AWS API,
but also may be due to implementation issues.
It should be avoided.

Resource Identity schemas are versioned starting from 0, so the first updated schema version will be version 1.
The current version of the Resource Identity schema is indicated by the annotation `@IdentityVersion(<version number>)`.

When updating the Resource Identity schema version, new acceptance tests are generated to ensure that existing resources can update their Resource Identity.
This is specified by adding one `@Testing(identityVersion="<identity-schema-version>;<provider-version>")` annotation per Resource Identity schema version.
See the [acceptance test generation documentation](acc-test-generation.md#adding-a-new-resource-identity-schema-version) for more details.

If a Resource Identity attribute is being added or renamed, the resource type will also need an Identity Upgrader.
Removing an attribute does not require an Identity Upgrader.

### Plugin Framework Upgrader

Identity Upgraders are currently not implemented for Plugin-Framework-based resource types, as there has not been need for it.
It is tracked by [GitHub Issue #44863](https://github.com/hashicorp/terraform-provider-aws/issues/44863).

### Plugin SDK Upgrader

In the Plugin SDK, each schema version defines a `schema.IdentityUpgrader` which upgrades the Resource Identity from a given version to the next.
The upgrade function has the signature `func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error)`.
By convention, the upgrader name is `<resource name>IdentityUpgradeV<source version>`.

For the version which is the target of the upgrader, add the parameter `sdkV2IdentityUpgraders="<function name>"` to the `@IdentityVersion` annotation.

For example, the Resource Identity for the resource type `aws_ec2_image_block_public_access` was updated to add the attribute `region` in version 6.21.0 of the provider.
The annotations are

```go
// @IdentityVersion(1, sdkV2IdentityUpgraders="imageBlockPublicAccessIdentityUpgradeV0")
// @Testing(identityVersion="0;v6.0.0")
// @Testing(identityVersion="1;v6.21.0")
```

The upgrader is

```go
var imageBlockPublicAccessIdentityUpgradeV0 = schema.IdentityUpgrader{
	Version: 0,
	Upgrade: func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		rawState[names.AttrRegion] = meta.(*conns.AWSClient).Region(ctx)
		return rawState, nil
	},
}
```
