---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry"
description: |-
  Manages an AWS Agent Registry registry.
---

# Resource: aws_agentregistry_registry

Manages an AWS Agent Registry registry.

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
      custom_jwt_authorizer {
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
}
```

## Argument Reference

The following arguments are required:

* `discovery_configuration` - (Required) Discovery configuration for the registry. [See below](#discovery_configuration-block).
* `name` - (Required) Name of the registry. Must start with a letter or digit. Valid characters are a-z, A-Z, 0-9, _ (underscore), - (hyphen), . (dot), and / (forward slash). The name can have up to 64 characters.

The following arguments are optional:

* `approval_configuration` - (Optional) Approval configuration for registry records. [See below](#approval_configuration-block).
* `auto_detection_configuration` - (Optional) Auto-detection configuration for the registry. When provided, the registry is automatically populated with resources discovered according to the configuration. [See below](#auto_detection_configuration-block).
* `description` - (Optional) Description of the registry. Maximum length of 4096 characters.
* `encryption_configuration` - (Optional) Server-side encryption configuration for the registry. [See below](#encryption_configuration-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `approval_configuration` Block

The `approval_configuration` configuration block supports the following arguments:

* `auto_approval_rules` - (Optional) Set of rules that determine which registry records are automatically approved on submission. Valid values: `APPROVE_ALL`. When omitted or empty, submitted records require manual review.

### `auto_detection_configuration` Block

The `auto_detection_configuration` configuration block supports the following arguments:

* `enabled` - (Required) Whether auto-detection is requested for the registry.
* `scope` - (Required) Source from which resources are detected. Valid values: `ORGANIZATION`.

### `discovery_configuration` Block

The `discovery_configuration` configuration block supports the following arguments:

* `authorizer_configuration` - (Optional) Authorizer configuration for the registry. Required when `authorizer_type` is `CUSTOM_JWT`. [See below](#authorizer_configuration-block).
* `authorizer_type` - (Required, Forces new resource) Type of authorizer that controls how consumers access the registry's search and MCP invoke operations. Valid values: `AWS_IAM`, `CUSTOM_JWT`.

### `authorizer_configuration` Block

The `authorizer_configuration` configuration block supports the following arguments:

* `custom_jwt_authorizer` - (Optional) Configuration for a custom JWT authorizer.

### `custom_jwt_authorizer` Block

The `custom_jwt_authorizer` configuration block supports the following arguments:

* `allowed_audience` - (Optional) Audience values accepted during JWT validation. A token is rejected if none of its audience claims match.
* `allowed_clients` - (Optional) Client identifiers accepted during JWT validation. A token is rejected if it was not issued to one of these clients.
* `allowed_scopes` - (Optional) Scopes accepted during JWT validation. A token is rejected if it does not carry one of these scopes.
* `custom_claim` - (Optional) Custom claims for additional JWT validation beyond standard OIDC claims. [See below](#custom_claim-block).
* `discovery_url` - (Required) OpenID Connect discovery URL used to retrieve the identity provider's metadata and signing keys.
* `private_endpoint` - (Optional) Private endpoint used to reach the identity provider's discovery URL over a private network path. [See below](#private_endpoint-block).
* `private_endpoint_override` - (Optional) Per-domain private endpoint overrides that route specific identity provider domains through distinct private endpoints. [See below](#private_endpoint_override-block).

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

### `private_endpoint` Block

Exactly one of the following must be specified:

* `managed_vpc_resource` - (Optional) Private endpoint backed by a service-managed VPC resource. [See below](#managed_vpc_resource-block).
* `self_managed_lattice_resource` - (Optional) Private endpoint backed by a self-managed VPC Lattice resource configuration. [See below](#self_managed_lattice_resource-block).

### `managed_vpc_resource` Block

* `endpoint_ip_address_type` - (Required) IP address type used by the private endpoint, either `IPV4` or `IPV6`.
* `routing_domain` - (Optional) Routing domain used to resolve traffic through the private endpoint.
* `security_group_ids` - (Optional) IDs of the security groups associated with the private endpoint network interfaces.
* `subnet_ids` - (Required) IDs of the subnets in which the private endpoint network interfaces are placed.
* `tags` - (Optional) Tags applied to the service-managed VPC resource.
* `vpc_identifier` - (Required) ID of the VPC in which the private endpoint is provisioned.

### `self_managed_lattice_resource` Block

* `resource_configuration_identifier` - (Required) Identifier of the VPC Lattice resource configuration, specified as a resource configuration ID or ARN.

### `private_endpoint_override` Block

* `domain` - (Required) Domain name to which this private endpoint override applies.
* `private_endpoint` - (Required) Private endpoint used to reach the specified domain. [See above](#private_endpoint-block).

### `encryption_configuration` Block

The `encryption_configuration` configuration block supports the following arguments:

* `kms_key_arn` - (Required) ARN of the customer-managed AWS KMS key used to encrypt the registry's content.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_arn` - ARN of the registry.
* `registry_id` - Unique identifier of the registry.
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

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
