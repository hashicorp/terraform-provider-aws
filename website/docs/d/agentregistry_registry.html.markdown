---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry"
description: |-
  Terraform data source for managing an AWS Agent Registry Registry.
---

# Data Source: aws_agentregistry_registry

Terraform data source for managing an AWS Agent Registry Registry.

## Example Usage

### Basic Usage

```terraform
data "aws_agentregistry_registry" "example" {
  registry_id = "registry-id-12345678"
}
```

### Lookup by ARN

```terraform
data "aws_agentregistry_registry" "example" {
  registry_id = "arn:aws:agent-registry:us-west-2:123456789012:registry/registry-id-12345678"
}
```

## Argument Reference

The following arguments are required:

* `registry_id` - (Required) Identifier of the registry to retrieve. Accepts a registry ID or ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `approval_configuration` - Approval configuration for registry records. [See below](#approval_configuration-block-attribute-reference).
* `created_at` - Timestamp when the registry was created.
* `description` - Description of the registry.
* `discovery_configuration` - Discovery configuration for the registry. [See below](#discovery_configuration-block-attribute-reference).
* `name` - Name of the registry.
* `registry_arn` - ARN of the registry.
* `status` - Current status of the registry. Valid values: `CREATING`, `READY`, `UPDATING`, `DELETING`, `CREATE_FAILED`, `UPDATE_FAILED`, `DELETE_FAILED`.
* `tags` - Map of tags assigned to the registry.
* `updated_at` - Timestamp when the registry was last updated.

### `approval_configuration` Block Attribute Reference

* `auto_approval_rules` - Set of rules that determine which registry records are automatically approved on submission. When empty, submitted records require manual review.

### `discovery_configuration` Block Attribute Reference

* `authorizer_configuration` - Authorizer configuration for the registry. Present when `authorizer_type` is `CUSTOM_JWT`. [See below](#authorizer_configuration-block-attribute-reference).
* `authorizer_type` - Type of authorizer that controls how consumers access the registry's search and MCP invoke operations. Valid values: `AWS_IAM`, `CUSTOM_JWT`.

### `authorizer_configuration` Block Attribute Reference

* `allowed_audience` - Audience values accepted during JWT validation.
* `allowed_clients` - Client identifiers accepted during JWT validation.
* `allowed_scopes` - Scopes accepted during JWT validation.
* `custom_claim` - Custom claims for additional JWT validation beyond standard OIDC claims. [See below](#custom_claim-block-attribute-reference).
* `discovery_url` - OpenID Connect discovery URL used to retrieve the identity provider's metadata and signing keys.

### `custom_claim` Block Attribute Reference

* `authorizing_claim_match_value` - Claim match criteria. [See below](#authorizing_claim_match_value-block-attribute-reference).
* `inbound_token_claim_name` - Name of the claim validated in the inbound JWT token.
* `inbound_token_claim_value_type` - Type of the claim value. Valid values: `STRING`, `STRING_ARRAY`.

### `authorizing_claim_match_value` Block Attribute Reference

* `claim_match_operator` - Operator used to match claim values. Valid values: `EQUALS`, `CONTAINS`, `CONTAINS_ANY`.
* `claim_match_value` - Value matched against. [See below](#claim_match_value-block-attribute-reference).

### `claim_match_value` Block Attribute Reference

* `match_value_string` - Single string value to match.
* `match_value_string_list` - Set of string values to match.
