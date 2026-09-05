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

      custom_claim {
        inbound_token_claim_name       = "sub"
        inbound_token_claim_value_type = "STRING"

        authorizing_claim_match_value {
          claim_match_operator = "EQUALS"

          claim_match_value {
            match_value_string = "authorized-user"
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `discovery_configuration` - (Required) Discovery configuration for the registry. [See below](#discovery_configuration-block).
* `name` - (Required) Name of the registry. Must contain only letters, numbers, hyphens, and underscores. Maximum length of 64 characters.

The following arguments are optional:

* `approval_configuration` - (Optional) Approval configuration for registry records. [See below](#approval_configuration-block).
* `description` - (Optional) Description of the registry. Maximum length of 4096 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `approval_configuration` Block

The `approval_configuration` configuration block supports the following arguments:

* `auto_approval_rules` - (Optional) Set of rules that determine which registry records are automatically approved on submission. Valid values: `APPROVE_ALL`. When omitted or empty, submitted records require manual review.

### `discovery_configuration` Block

The `discovery_configuration` configuration block supports the following arguments:

* `authorizer_configuration` - (Optional) Authorizer configuration for the registry. Required when `authorizer_type` is `CUSTOM_JWT`. [See below](#authorizer_configuration-block).
* `authorizer_type` - (Required, Forces new resource) Type of authorizer that controls how consumers access the registry's search and MCP invoke operations. Valid values: `AWS_IAM`, `CUSTOM_JWT`.

### `authorizer_configuration` Block

The `authorizer_configuration` configuration block supports the following arguments:

* `allowed_audience` - (Optional) Audience values accepted during JWT validation. A token is rejected if none of its audience claims match.
* `allowed_clients` - (Optional) Client identifiers accepted during JWT validation. A token is rejected if it was not issued to one of these clients.
* `allowed_scopes` - (Optional) Scopes accepted during JWT validation. A token is rejected if it does not carry one of these scopes.
* `custom_claim` - (Optional) Custom claims for additional JWT validation beyond standard OIDC claims. [See below](#custom_claim-block).
* `discovery_url` - (Required) OpenID Connect discovery URL used to retrieve the identity provider's metadata and signing keys.

### `custom_claim` Block

The `custom_claim` configuration block supports the following arguments:

* `authorizing_claim_match_value` - (Required) Claim match criteria. [See below](#authorizing_claim_match_value-block).
* `inbound_token_claim_name` - (Required) Name of the claim to validate in the inbound JWT token. Must contain only letters, numbers, and the characters `_`, `.`, `-`, `:`.
* `inbound_token_claim_value_type` - (Required) Type of the claim value. Valid values: `STRING`, `STRING_ARRAY`.

### `authorizing_claim_match_value` Block

The `authorizing_claim_match_value` configuration block supports the following arguments:

* `claim_match_operator` - (Required) Operator used to match claim values. Valid values: `EQUALS`, `CONTAINS`, `CONTAINS_ANY`.
* `claim_match_value` - (Required) Value to match against. [See below](#claim_match_value-block).

### `claim_match_value` Block

The `claim_match_value` configuration block supports exactly one of the following arguments:

* `match_value_string` - (Optional) Single string value to match. Must contain only letters, numbers, and the characters `_`, `.`, `-`, `:`.
* `match_value_string_list` - (Optional) Set of string values to match. Each value must contain only letters, numbers, and the characters `_`, `.`, `-`, `:`.

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

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_agentregistry_registry.example
  identity = {
    registry_id = "registry-id-12345678"
  }
}

resource "aws_agentregistry_registry" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `registry_id` (String) Registry ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Agent Registry Registry by registry ID. For example:

```terraform
import {
  to = aws_agentregistry_registry.example
  id = "registry-id-12345678"
}
```

Using `terraform import`, import an Agent Registry Registry by registry ID. For example:

```console
% terraform import aws_agentregistry_registry.example registry-id-12345678
```
