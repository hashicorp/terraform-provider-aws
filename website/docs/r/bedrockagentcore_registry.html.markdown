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

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Optional) JWT-based authorization configuration block. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer` Block

The `custom_jwt_authorizer` block supports the following:

* `discovery_url` - (Required) URL used to fetch OpenID Connect configuration or authorization server metadata. Must end with `.well-known/openid-configuration`.
* `allowed_audience` - (Optional) Set of allowed audience values for JWT token validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT token validation.
* `allowed_scopes` - (Optional) Set of scopes that are allowed to access the token.
* `allowed_workload_configuration` - (Optional) Configuration restricting which workloads may use this authorizer. See [`allowed_workload_configuration`](#allowed_workload_configuration) below.
* `custom_claim` - (Optional) Repeatable block to define a custom claim validation name, value, and operation. See [`custom_claim`](#custom_claim) below.
* `private_endpoint` - (Optional) Private endpoint used to reach the authorization server. See [`private_endpoint`](#private_endpoint) below.
* `private_endpoint_overrides` - (Optional) Overrides for the private endpoints used to reach the authorization server. See [`private_endpoint_overrides`](#private_endpoint_overrides) below.

### `allowed_workload_configuration` Block

* `hosting_environment` - (Optional) Hosting environments allowed to use the authorizer. Between 1 and 10 entries. See [`hosting_environment`](#hosting_environment) below.
* `workload_identities` - (Optional) List of workload identity names allowed to use the authorizer. Between 1 and 10 entries.

### `hosting_environment` Block

* `arn` - (Required) ARN of the hosting environment.

### `private_endpoint_overrides` Block

* `domain` - (Required) Domain the override applies to.
* `private_endpoint` - (Required) Private endpoint configuration. See [`private_endpoint`](#private_endpoint) below.

### `private_endpoint` Block

Exactly one of the following must be specified:

* `managed_vpc_resource` - (Optional) Managed VPC resource configuration. See [`managed_vpc_resource`](#managed_vpc_resource) below.
* `self_managed_lattice_resource` - (Optional) Self-managed VPC Lattice resource configuration. See [`self_managed_lattice_resource`](#self_managed_lattice_resource) below.

### `managed_vpc_resource` Block

* `endpoint_ip_address_type` - (Required) IP address type for the endpoint. Valid values are `IPV4` and `IPV6`.
* `subnet_ids` - (Required) IDs of the subnets for the endpoint.
* `vpc_identifier` - (Required) Identifier of the VPC for the endpoint.
* `routing_domain` - (Optional) Routing domain for the endpoint.
* `security_group_ids` - (Optional) IDs of the security groups for the endpoint.
* `tags` - (Optional) Tags to assign to the managed VPC resource.

### `self_managed_lattice_resource` Block

* `resource_configuration_identifier` - (Required) Identifier of the VPC Lattice resource configuration.

### `custom_claim` Block

The `custom_claim` block supports the following:

* `authorizing_claim_match_value` - (Required) Configuration block to define the value or values to match for and the relationship of the match. See [`authorizing_claim_match_value`](#authorizing_claim_match_value) below.
* `inbound_token_claim_name` - (Required) Name of the custom claim field to check.
* `inbound_token_claim_value_type` - (Required) Data type of the claim value to check for. Valid values are `STRING` and `STRING_ARRAY`.

### `authorizing_claim_match_value` Block

The `authorizing_claim_match_value` block supports the following:

* `claim_match_operator` - (Required) Relationship between the claim field value and the value or values to match for. Valid values are `EQUALS`, `CONTAINS`, and `CONTAINS_ANY`. `EQUALS` can be used only when `inbound_token_claim_value_type` is `STRING`. `CONTAINS` or `CONTAINS_ANY` can be used only when `inbound_token_claim_value_type` is `STRING_ARRAY`.
* `claim_match_value` - (Required) Value or values to match for. See [`claim_match_value`](#claim_match_value) below.

### `claim_match_value` Block

The `claim_match_value` block supports the following:

* `match_value_string` - (Optional) String value to match for. Must be specified when `claim_match_operator` is `EQUALS` or `CONTAINS`. Exactly one of `match_value_string` or `match_value_string_list` must be specified.
* `match_value_string_list` - (Optional) List of strings to check for a match. Must be specified when `claim_match_operator` is `CONTAINS_ANY`. Exactly one of `match_value_string` or `match_value_string_list` must be specified.

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
