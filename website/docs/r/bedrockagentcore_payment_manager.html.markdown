---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_payment_manager"
description: |-
  Manages an AWS Bedrock AgentCore Payment Manager.
---

# Resource: aws_bedrockagentcore_payment_manager

Manages an AWS Bedrock AgentCore Payment Manager. A Payment Manager governs how agents authenticate and authorize payment operations through AgentCore.

## Example Usage

### AWS IAM Authorization

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example-payment-manager"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_bedrockagentcore_payment_manager" "example" {
  name            = "example-payment-manager"
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.example.arn
}
```

### Custom JWT Authorization

```terraform
resource "aws_bedrockagentcore_payment_manager" "example" {
  name            = "example-payment-manager"
  authorizer_type = "CUSTOM_JWT"
  role_arn        = aws_iam_role.example.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      allowed_audience = ["example-audience"]
      discovery_url    = "https://example.com/.well-known/openid-configuration"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `authorizer_type` - (Required) Type of authorizer used by the payment manager. Valid values: `AWS_IAM`, `CUSTOM_JWT`.
* `name` - (Required, Forces new resource) Name of the payment manager.
* `role_arn` - (Required) ARN of the IAM role that the payment manager assumes.

The following arguments are optional:

* `authorizer_configuration` - (Optional) Configuration for the authorizer. Required when `authorizer_type` is `CUSTOM_JWT`. See [`authorizer_configuration`](#authorizer_configuration) below.
* `description` - (Optional) Description of the payment manager.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authorizer_configuration` Block

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Required) JWT-based authorization configuration block. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer` Block

The `custom_jwt_authorizer` block supports the following:

* `allowed_audience` - (Optional) Set of allowed audience values for JWT token validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT token validation.
* `allowed_scopes` - (Optional) Set of scopes that are allowed to access the token.
* `allowed_workload_configuration` - (Optional) Configuration restricting which workloads may use this authorizer. See [`allowed_workload_configuration`](#allowed_workload_configuration) below.
* `custom_claim` - (Optional) Repeatable block to define a custom claim validation name, value, and operation. See [`custom_claim`](#custom_claim) below.
* `discovery_url` - (Required) URL used to fetch OpenID Connect configuration or authorization server metadata. Must end with `.well-known/openid-configuration`.
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
* `routing_domain` - (Optional) Routing domain for the endpoint.
* `security_group_ids` - (Optional) IDs of the security groups for the endpoint.
* `subnet_ids` - (Required) IDs of the subnets for the endpoint.
* `tags` - (Optional) Tags to assign to the managed VPC resource.
* `vpc_identifier` - (Required) Identifier of the VPC for the endpoint.

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

* `payment_manager_arn` - ARN of the Payment Manager.
* `payment_manager_id` - Unique identifier of the Payment Manager.
* `status` - Status of the Payment Manager.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `workload_identity_details` - Workload identity details for the Payment Manager.
    * `workload_identity_arn` - ARN of the workload identity.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_manager.example
  identity = {
    payment_manager_id = "payment-manager-id-12345678"
  }
}

resource "aws_bedrockagentcore_payment_manager" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `payment_manager_id` (String) Payment manager ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Payment Manager by payment manager ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_manager.example
  id = "payment-manager-id-12345678"
}
```

Using `terraform import`, import a Bedrock AgentCore Payment Manager by payment manager ID. For example:

```console
% terraform import aws_bedrockagentcore_payment_manager.example payment-manager-id-12345678
```
