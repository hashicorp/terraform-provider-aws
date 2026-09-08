---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry"
description: |-
  Terraform resource for managing an AWS Agent Registry Registry.
---

# Resource: aws_agentregistry_registry

Terraform resource for managing an AWS Agent Registry Registry.

A registry allows developers to discover, manage, and govern reusable agentic components such as tools, prompts, guardrails, and knowledge bases.

## Example Usage

### Basic Usage

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
```

### With Description

```terraform
resource "aws_agentregistry_registry" "example" {
  name        = "example-registry"
  description = "Example agent registry"

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
```

### With Auto Approval

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  approval_configuration {
    auto_approval_rules = ["APPROVE_ALL"]
  }

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
```

### With Custom JWT Authorization

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      discovery_url    = "https://example.com/.well-known/openid-configuration"
      allowed_audience = ["https://api.example.com"]
      allowed_clients  = ["client-id-1"]
      allowed_scopes   = ["read", "write"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the registry. Must contain only letters, numbers, hyphens, and underscores. Maximum length of 64 characters.
* `discovery_configuration` - (Required) Discovery configuration for the registry. See [`discovery_configuration`](#discovery_configuration) below.

The following arguments are optional:

* `approval_configuration` - (Optional) Approval configuration for registry records. See [`approval_configuration`](#approval_configuration) below.
* `description` - (Optional) Description of the registry. Maximum length of 4096 characters.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `approval_configuration`

* `auto_approval_rules` - (Optional) Set of rules that determine which registry records are automatically approved on submission. Valid values: `APPROVE_ALL`. When omitted or empty, submitted records require manual review.

### `discovery_configuration`

* `authorizer_type` - (Required) Type of authorizer that controls how consumers access the registry's search and MCP invoke operations. Valid values: `AWS_IAM`, `CUSTOM_JWT`. **Changing this value will recreate the resource.**
* `authorizer_configuration` - (Optional) Authorizer configuration for the registry. Required when `authorizer_type` is `CUSTOM_JWT`. See [`authorizer_configuration`](#authorizer_configuration) below.

### `authorizer_configuration`

* `discovery_url` - (Required) OpenID Connect discovery URL used to retrieve the identity provider's metadata and signing keys.
* `allowed_audience` - (Optional) Audience values accepted during JWT validation. A token is rejected if none of its audience claims match.
* `allowed_clients` - (Optional) Client identifiers accepted during JWT validation. A token is rejected if it was not issued to one of these clients.
* `allowed_scopes` - (Optional) Scopes accepted during JWT validation. A token is rejected if it does not carry one of these scopes.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_arn` - ARN of the registry.
* `registry_id` - Unique identifier of the registry.
* `status` - Current status of the registry. Valid values: `CREATING`, `READY`, `UPDATING`, `DELETING`, `CREATE_FAILED`, `UPDATE_FAILED`, `DELETE_FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agent Registry Registry using the `registry_id`. For example:

```terraform
import {
  to = aws_agentregistry_registry.example
  id = "AIDACKCEVSQ6C2EXAMPLE"
}
```

Using `terraform import`, import Agent Registry Registry using the `registry_id`. For example:

```console
% terraform import aws_agentregistry_registry.example AIDACKCEVSQ6C2EXAMPLE
```
