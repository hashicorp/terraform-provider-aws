<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Supporting module-scoped User-Agents with Terraform Protocol V5

**Summary:** In order to implement `provider_meta` in both the standard and AWSCC provider, we need to navigate the absence of support for structural attributes in Terraform Plugin SDK V2 and Terraform protocol V5.  
**Created**: 2025-12-08  
**Author**: [@jar-b](https://github.com/jar-b)  

---

The standard and AWSCC provider are actively working on adding support for [`provider_meta`](https://developer.hashicorp.com/terraform/internals/provider-meta) to allow module authors to supply additional values to the User-Agent header sent to AWS APIs during CRUD operations. Because the standard provider is [muxed](https://developer.hashicorp.com/terraform/plugin/mux) and [still supports Terraform Protocol V5](https://github.com/hashicorp/terraform-provider-aws/blob/v6.16.0/main.go#L29), we are unable to use a “list of objects” attribute in the `provider_meta` schema for this provider.

In order to proceed, we need to align on a schema definition that can function within the current operational constraints of both providers.

## Background

The AWSCC provider already has a [working implementation](https://github.com/hashicorp/terraform-provider-awscc/pull/474) of `provider_meta` support in which the schema contains a single `user_agent` attribute that is a list of objects. This schema matches the existing provider-level configuration [attribute](https://registry.terraform.io/providers/hashicorp/awscc/latest/docs#custom-user-agent-information) of the same name. Notably, the AWSCC provider uses only [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) and operates with Terraform Protocol V6.

```
  provider_meta "awscc" {
    user_agent = [
      {
        product_name    = "jb-demo"
        product_version = "0.0.1"
        comment         = "a demo module"
      },
    ]
  }
```

In an ideal scenario, the standard provider would match the rough shape of the schema, but replace list attributes with blocks.

```
  provider_meta "aws" {
    user_agent {
      product_name    = "jb-demo"
      product_version = "0.0.1"
      comment         = "a demo module"
    }
  }
```

However, the schema cannot be written in this way as Terraform Protocol V5 doesn't support structural attributes, meaning an attribute with this shape can't be defined in the protocol at all.

```
│ Error: Failed to load plugin schemas
│
│ Error while loading schemas for plugin components: Failed to obtain provider schema: Could not load the schema for provider registry.terraform.io/hashicorp/aws: failed to retrieve
│ schema from provider "registry.terraform.io/hashicorp/aws": Error converting provider_meta schema: The provider_meta schema couldn't be converted into a usable type. This is always
│ a problem with the provider. Please report the following to the provider developer:
│
│ AttributeName("user_agent"): protocol version 5 cannot have Attributes set..
```

Additionally, even if the provider were upgraded to Terraform Protocol V6, Terraform Plugin Framework lacks support for blocks in the [`metaschema.Schema`](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.16.1/provider/metaschema/schema.go#L20-L31) struct, making it impossible to match the schema shape between the Plugin SDK V2 and Plugin Framework providers as [required by](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-mux@v0.21.0/tf5muxserver#NewMuxServer) `terraform-plugin-mux`.[^1]

Assuming that changes to the upstream plugin libraries are off the table due to timing constraints, we’re left with options that require some form of a trade off on the provider side.

### Option 1 - Simplify the `provider_meta` Schema

This option would involve simplifying the `provider_meta` schema for **both** providers such that `user_agent` is a list of string type. only a list of string argument.

```
  provider_meta "aws" {
    user_agent = [
      "jb-demo/0.0.1 (a demo module)"
    ]
  }
```

Rather than composing the product name, version, and comment into a User-Agent entry on the user’s behalf, module authors would be responsible for formatting the additional content themselves. Inclusion of a provider-defined function to compose a UA string from its constituent parts can offset the impact of this change.

The primary drawback to this approach is that the AWSCC provider already supports a provider-level [`user_agent` argument](https://registry.terraform.io/providers/hashicorp/awscc/latest/docs#custom-user-agent-information) with the complex schema. We could opt to deprecate the complex form and change it in a subsequent major version for consistency, but inconsistency will be present for at least some time with this approach.

### Option 2 - Simplify the `provider_meta` Schema for Standard Only

Similar to option 1, but leaves the AWSCC implementation unchanged. The standard and AWSCC provider will have different `provider_meta` schema definitions for the foreseeable future.

## Decision

We will pursue option 1 and utilize a simplified version on the schema for the `user_agent` attribute in both the `provider_meta` and `provider` configurations. The AWSCC provider will retain the complex `user_agent` provider argument for the duration of the `v1.X` series, and change over to the simplified version in `v2.0.0` as a breaking change.

## Consequences/Future Work

- The existing AWSCC `provider_meta` [PR](https://github.com/hashicorp/terraform-provider-awscc/pull/474) will need to be refactored to utilize the simplified schema.
- We will need to create an issue to track the breaking change to `user_agent` slated for `v2.0.0`.  
- In `hashicorp/aws-sdk-go-base` , the [`useragent` package](https://github.com/hashicorp/aws-sdk-go-base/tree/v2.0.0-beta.68/useragent) and corresponding types from `internal/config` should be simplified from structs to a list of strings. This can be done after V2 of AWSCC is released and the explicit structure is no longer necessary.
- The PRs implementing `provider_meta` should include a `build_user_agent` provider-defined function to facilitate proper formatting of argument values.

[^1]:  This particular limitation means that the standard provider will effectively never be able to implement a structural attribute in `provider_meta`. Eliminating muxing from the provider would require thousands of Plugin SDK V2 based resources to be migrated to Plugin Framework. There is no currently available tooling which makes this feasible at the scale of the AWS provider.
