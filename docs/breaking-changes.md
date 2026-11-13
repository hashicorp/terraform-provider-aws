<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Breaking Changes

A breaking change to the provider is any change that requires an end user to modify a previously valid Terraform configuration in order to maintain an existing deployment. This usually means that `terraform plan` will show no unexpected diffs after upgrading the provider.

Breaking changes are not allowed within a provider major version.

## What's A Breaking Change

- Removing a resource, data source, ephemeral resource, list resource or provider function.
- Removing an attribute.
- Renaming an attribute without supporting the previous attribute name.
- Making an Optional attribute Required.
- Removing Computed from an attribute.
- Making attribute validation more restrictive.
- Changing the default value of an attribute.
- Any change that causes `terraform plan` to show unexpected diffs between minor version upgrades.
- Any change to how a resource is created, updated, or imported that alters expected behavior.

## What's Not A Breaking Change

- Adding a resource, data source, ephemeral resource, list resource or provider function.
- Adding an Optional or Computed-only attribute.
- Making attribute validation less restrictive (e.g. adding new valid values to a String attribute).
- Bug fixes that correct behavior to match authoritative documentation.

## Guidelines

- Breaking changes require a provider major release.
- Deprecate before removing. Authoritative documentation:
    - [Terraform Plugin SDK v2](https://developer.hashicorp.com/terraform/plugin/sdkv2/best-practices/deprecations)
    - [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework/deprecations)
- No unexpected `terraform plan` diffs on minor or patch version upgrades.
