<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# When Not to Support `List` on a Resource Type

## Background

One of the primary use cases for List Resources is to allow practitioners to enumerate and import remote resources into Terraform management.
This means that, in general, there should be a List Resource corresponding to a Resource type.
This includes Singleton cases where there is a single remote resource in a given AWS region or account, so that the remote resource can be identified and imported.

There are some rare cases where a corresponding List Resource should not be created.

## Excluded Resource Patterns

### Exclusive Resources

Exclusive Resources, described in [Exclusive Relationship Management Resources](./exclusive-relationship-management-resources.md), are not importable,
as they enforce a stronger restriction than standard resource types.
See [Do Not Implement Import Support for Exclusive Resource Types](./no-import-support-for-exclusive-resource-types.md) for more details.

As they are not importable, they will not support List.
The List operation for the standard, non-exclusive resource type can be used to List and import the remote resources.

### Waiter Resources

A **waiter resource** superficially resembles a **property entity** as it has a 1:1 or 1:[0,1] relationship with an associated resource.
However, it exists only to allow the provider to carry out a multi-step process and wait for or create dependencies up completion.

As there is no corresponding remote resource, there is nothing to List or Import.

## Excluded Resource Types

As new excluded resource types are identified, they should be added to this list.

### Exclusive Resources

* `aws_iam_policy_attachment`

### Waiter Resources

* `aws_acm_certificate_validation`
