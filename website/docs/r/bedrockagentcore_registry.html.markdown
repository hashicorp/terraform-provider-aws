---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry"
description: |-
  Manages an AWS Bedrock AgentCore Registry.
---

# Resource: aws_bedrockagentcore_registry

Manages an AWS Bedrock AgentCore Registry. A registry serves as a centralized catalog for organizing and managing registry records, including MCP servers, A2A agents, agent skills, and custom resource types.

!> **Warning:** This resource is deprecated. AWS Agent Registry is currently available in public preview. [On August 6, 2026]((https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-faq.html#registry-faq-what-is-changing)) functionality will move from the `bedrock-agentcore` namespace to the `agent-registry` namespace. This resource will continue to work until [September 17, 2026](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-faq.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_registry" "example" {
  name = "example_registry"
}
```

### With Description and Auto Approval

```terraform
resource "aws_bedrockagentcore_registry" "example" {
  name          = "example_registry"
  description   = "MCP servers and tools for the platform team"
  auto_approval = true
}
```

### With Custom JWT Authorizer

```terraform
resource "aws_bedrockagentcore_registry" "example" {
  name            = "example_registry"
  authorizer_type = "CUSTOM_JWT"

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://example.okta.com/.well-known/openid-configuration"
      allowed_audience = ["audience-id"]
      allowed_clients  = ["client-id"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the registry. Must be unique within your account and contain only letters, numbers, hyphens, and underscores. Maximum length of 64 characters.

The following arguments are optional:

* `approval_configuration` - (Optional)  Approval configuration for registry records. [See below](#approval_configuration-block).
* `authorizer_configuration` - (Optional) Authorizer configuration for the registry. Required when `authorizer_type` is `CUSTOM_JWT`. [See below](#authorizer_configuration-block).
* `authorizer_type` - (Optional, Forces new resource) Type of authorizer to use for the registry. Valid values are `AWS_IAM` (default) and `CUSTOM_JWT`. This controls the authorization method for the Search and Invoke APIs used by consumers.
* `description` - (Optional) Description of the registry.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `approval_configuration` Block

The `approval_configuration` configuration block supports the following arguments:

* `auto_approval` - (Optional) Whether registry records are auto-approved. When set to `true`, records are automatically approved upon creation. When set to `false` (the default), records require explicit approval.

### `authorizer_configuration` Block

The `authorizer_configuration` configuration block supports the following arguments:

* `custom_jwt_authorizer` - (Optional) Inbound JWT-based authorization configuration. [See below](#custom_jwt_authorizer-block).

### `custom_jwt_authorizer` Block

The `custom_jwt_authorizer` configuration block supports the following arguments:

* `private_endpoint` - (Optional) Private endpoint used to reach the OIDC discovery URL over a VPC. See [`private_endpoint`](#private_endpoint) below.
* `private_endpoint_overrides` - (Optional) Up to 5 per-domain private endpoint overrides. See [`private_endpoint_overrides`](#private_endpoint_overrides) below.
* `allowed_audience` - (Optional) Set of audience values validated during incoming JWT token validation.
* `allowed_clients` - (Optional) Set of client IDs validated during incoming JWT token validation.
* `allowed_scopes` - (Optional) Set of scopes allowed to access the token.
* `advertised_scope_mapping` - (Optional) Map of up to 50 entries associating each scope in `allowed_scopes` with the scope value advertised in OAuth protected-resource metadata. Use when the scope clients request from the identity provider differs from the scope in the validated token.
* `custom_claim` - (Optional) Set of custom claim validations. [See below](#custom_claim-block).
* `discovery_url` - (Required) OpenID Connect discovery URL used to validate incoming tokens.

### `custom_claim` Block

The `custom_claim` configuration block supports the following arguments:

* `authorizing_claim_match_value` - (Required) Match value configuration for the claim. [See below](#authorizing_claim_match_value-block).
* `inbound_token_claim_name` - (Required) Name of the inbound token claim to validate.
* `inbound_token_claim_value_type` - (Required) Value type of the inbound token claim.

### `authorizing_claim_match_value` Block

The `authorizing_claim_match_value` configuration block supports the following arguments:

* `claim_match_operator` - (Required) Relationship between the claim field value and the value or values you are matching for.
* `claim_match_value` - (Required) Value or values to match for. [See below](#claim_match_value-block).

### `claim_match_value` Block

The `claim_match_value` configuration block supports the following arguments. Exactly one of `match_value_string` or `match_value_string_list` must be set.

* `match_value_string` - (Optional) Single value to match for.
* `match_value_string_list` - (Optional) Set of values to match for.

### `private_endpoint`

The `private_endpoint` block supports exactly one of:

* `managed_vpc_resource` - (Optional) Service-managed VPC endpoint. See [`managed_vpc_resource`](#managed_vpc_resource) below.
* `self_managed_lattice_resource` - (Optional) Self-managed VPC Lattice resource. See [`self_managed_lattice_resource`](#self_managed_lattice_resource) below.

### `managed_vpc_resource`

The `managed_vpc_resource` block supports the following:

* `endpoint_ip_address_type` - (Required) IP address type for the endpoint. Valid values: `IPV4`, `IPV6`.
* `subnet_ids` - (Required) Set of subnet IDs for the endpoint.
* `vpc_identifier` - (Required) VPC identifier.
* `routing_domain` - (Optional) Routing domain for the endpoint.
* `security_group_ids` - (Optional) Set of up to 5 security group IDs.
* `tags` - (Optional) Map of tags for the endpoint.

### `self_managed_lattice_resource`

The `self_managed_lattice_resource` block supports the following:

* `resource_configuration_identifier` - (Required) VPC Lattice resource configuration ID or ARN.

### `private_endpoint_overrides`

The `private_endpoint_overrides` block supports the following:

* `domain` - (Required) Domain the override applies to.
* `private_endpoint` - (Required) Private endpoint for this domain. See [`private_endpoint`](#private_endpoint) above.



## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_arn` - ARN of the registry.
* `registry_id` - Unique identifier of the registry.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_registry.example
  identity = {
    registry_id = "registry-id-12345678"
  }
}

resource "aws_bedrockagentcore_registry" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `registry_id` (String) Registry ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Registry by registry ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_registry.example
  id = "registry-id-12345678"
}
```

Using `terraform import`, import a Bedrock AgentCore Registry by registry ID. For example:

```console
% terraform import aws_bedrockagentcore_registry.example registry-id-12345678
```
