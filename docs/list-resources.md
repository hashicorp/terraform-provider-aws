<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Enabling List Resource Support

Terraform version 1.14 introduced resource queries using the `list` block.

## Concepts

In order to implement the List Resource for a remote resource type,
the remote resource type must be categorized on several axes:
relationship to other resources,
how the AWS API works, and
Terraform considerations.

### Relationship to Other Resources

The relationship to other resources influences the design of the List Resource, as well as the resource type's Resource Identity.

This section will borrow terms from the [Entity-Relationship model](https://en.wikipedia.org/wiki/Entity–relationship_model) for database design.
The differentiating factor for List Resource implementation is that it is solely based on identifiers used for reading the remote resource,
rather than any relationship as in databases.

#### Strong Entity

A **strong entity** is a resource type that can be read or listed using only its own identity information.
It is still a strong entity if it must be associated with another resource, if that other resource's identifier is not needed for reading or listing.

For example, an EC2 Instance (`aws_instance`) can be read using only its `InstanceID`, so it is a **strong entity**, even though it must reside in a VPC Subnet, which is contained in a VPC.

#### Weak Entity

A **weak entity** is a resource type that cannot be read or listed using only its own identity information, and also requires the identifier of an associated resource.

For example, an ELB Listener (`aws_lb_listener`) can only be returned in a list operation by including the containing ELB Load Balancer's ARN, so it is a **weak entity**.

#### Property Entity

A property entity is a resource type that models a property of an associated resource in a 1:1 or 1:[0,1] relationship.
Property entities with a 1:[0,1] relationship can be called an **optional property entity**.
This is a special case of a **weak entity**.
Note that this is not a concept from the Entity-Relationship model.

For example, the S3 Bucket (`aws_s3_bucket`) has many properties modelled as separate resources, such as Bucket ACLs (`aws_s3_bucket_acl`), and Bucket Policies (`aws_s3_bucket_policy`).
A Bucket Policy is only created if the user requests one, so it is an **optional property entity**.

### AWS API Patterns

The AWS APIs used to list remote resources follow several patterns.

#### All-Or-One

The AWS request can be used either to retrieve a single remote resource when an identifier is supplied, or all remote resources when no identifier is supplied.

#### All-Or-Some

The AWS request can be used either to retrieve a set of remote resource when a set of identifiers is supplied, or all remote resources when no identifiers are supplied.

#### Summary List

The AWS API responds with a subset of remote resource information.
Subsequent API calls are required to retrieve full resource information.

A simple, very common example of this is when resource tags are not returned by the list call,
and they must be retrieved using a separate API call.

### Terraform Considerations

In some cases, how the AWS provider manages a resource type will influence the implementation of the List Resource.

#### Manageable Resources

A **manageable resource** is a remote resource than _can_ be brought into Terraform management.
In most cases, all remote resources of a resource type are **manageable resources**.
However, in some cases, AWS defines a set of default instances that cannot be modified by users.

For example, RDS and other RDS-derived services define a set of default Parameter Groups, which represent a set of datastore configuration parameters.
Users can create custom Parameter Groups but cannot make any changes to the defaults.
These default parameter groups would be **non-manageable resources**, and therefore should be neither importable nor listable.

#### "Special Case" Resource Types

Some resource types have associated resource types to allow for special handling, especially around creation and possibly deletion.

For example, the EC2 service defines default variants of `aws_vpc`, `aws_security_group`, `aws_network_acl`, `aws_route_table`, `aws_subnet`, and `aws_vpc_dhcp_options`.
These all implement the Adopt-on-Create pattern, and either have a no-op delete or ignore deletion failures.

There is typically a way to identify the remote resources as needing to be managed as a **“special case” resource type**, such as a flag or other value on the resource data.

## Implementation

By default, the List Resource must only return the Resource Identity and a Display Name for each remote resource.
If the parameter `IncludeResource` on the list request is set to `true`, the resource data should also be populated.

The Display Name should be set to a value that can uniquely identify the resource in a user-friendly way, as much as possible.
For resource types that have a `Name` field, or a name-equivalent such as` DBIdentifier` for RDS resources types, this is the best choice.
Some services, such as EC2 and ELB, treat a resource tag with the key `Name` as the name value in the AWS Console for many resource types.
In these cases, the value of the `Name` tag should be used.

**Non-manageable resources** should be excluded from the results, as they cannot be brought under Terraform management.

Because there is no mechanism for overriding the resource type of a returned resource, **“special case” resource types** should be excluded from the results.
This should be documented in practitioner documentation.
A List Resource for the **“special case” resource type** should be created.

If a resource type is a **weak entity**, but _not_ a **property entity**,
the List Resource schema should define a required attribute corresponding to the identifier of the parent resource.
For example, the `aws_s3_object` is a **weak entity** which requires the `bucket` attribute to identify the containing S3 Bucket.

### Plugin Framework

The scaffolding can be generated by the `skaff` tool in the desired service directory.

```console
skaff list --framework --name <resource-name>
```

The annotation `@FrameworkListResource("<resource_name>")` is required to register the List Resource with the provider.
The value of `<resource_name>` must match the name of the associated resource type.

When adding a List Resource for an existing resource type,
extract the portion of the existing resource type's Read operation that flattens the API response into the resource data model into a new method `flatten`.
For many resource types, this will simply call `flex.Flatten(...)`.

Both the List Resource and the resource type's Read operation should call the `flatten` function.

### Plugin SDK

The scaffolding can be generated by the `skaff` tool in the desired service directory.

```console
skaff list --name <resource-name>
```

The annotation `@SDKListResource("<resource_name>")` is required to register the List Resource with the provider.
The value of `<resource_name>` must match the name of the associated resource type.

In the iterator loop body,
the resource data `rd` must always be populated with the `id` value, using `rd.SetId(<value>)`.
Any additional attributes needed for the Resource Identity must be set using `rd.Set(<attribute-name>, <attribute-value>)`.

If the list request parameter `IncludeResource` is set, the resource data should be populated.
This should be done using a function named `resource<Resource Name>Flatten`.
Both the List Resource and the resource's Read operation should use this flatten function.
If the function does not exist, refactor the resource's Read operation so that the body of the function that sets values on the resource data is moved to the flattening function.

## Acceptance Testing

The `skaff` tool will generate scaffolding for acceptance tests for the List Resource.

### Tests to Create

List Resources must have the following tests:

* A `basic` test that validates that multiple resources can be queried
* An `includeResource` test that validates that resource data is populated when `IncludeResource` is set to `true`

List Resources for regional resources, which is most resource types, must also have:

* A `regionOverride` test that validates that the `region` attribute on the `list` block overrides the default region of the provider.

In general, the configuration should be as simple as possible, avoiding optional attributes.
One exception is if a resource type has an optional `name` attribute or the combination of `name` and `name_prefix`.
In that case, `name` should be specified if the `name` value is used in the Resource Identity.

If a resource type has attribute settings that could affect listing behavior,
additional tests to validate that behavior should be added.

### Test Configuration

All tests should create multiple instances of the resource type being tested.
Set the [`count` meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/count) to `var.resource_count`.
Other resources in the configuration should have a single instance, unless more than one is needed, for example EC2 Subnets when testing ELB Load Balancers.

For **property entities**, multiple instances of the parent resource must be created and each **property entity** resource must be associated with an instance of the parent resource.

The `includeResource` test should set the `tags` attribute if the resource type supports tagging.
Tags should only be applied to the resource that is being tested.
The `QueryResultChecks` should include a `querycheck.ExpectResourceKnownValues` check that verifies that each attribute on the resource has a value set to an expected value.
In the list query configuration file, the `include_resource` attribute should be set to `true`.

The `regionOverride` test should set the `region` attribute for all non-global resources in the configuration.
In the list query configuration file, the `region` attribute in the `config` block should be set to `var.region`.
