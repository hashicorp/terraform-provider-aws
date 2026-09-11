---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry"
description: |-
  Provides details about an AWS Agent Registry registry.
---

# Data Source: aws_agentregistry_registry

Provides details about an AWS Agent Registry registry.

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

* `approval_configuration` - Approval configuration for registry records. [See below](#approval_configuration-block).
* `created_at` - Timestamp when the registry was created.
* `description` - Description of the registry.
* `discovery_configuration` - Discovery configuration for the registry. [See below](#discovery_configuration-block).
* `encryption_configuration` - Server-side encryption configuration for the registry. [See below](#encryption_configuration-block).
* `name` - Name of the registry.
* `registry_arn` - ARN of the registry.
* `status` - Current status of the registry. Valid values: `CREATING`, `READY`, `UPDATING`, `DELETING`, `CREATE_FAILED`, `UPDATE_FAILED`, `DELETE_FAILED`.
* `tags` - Map of tags assigned to the registry.
* `updated_at` - Timestamp when the registry was last updated.

### `approval_configuration` Block

* `auto_approval_rules` - Set of rules that determine which registry records are automatically approved on submission. When empty, submitted records require manual review.

### `discovery_configuration` Block

* `authorizer_configuration` - Authorizer configuration for the registry. Present when `authorizer_type` is `CUSTOM_JWT`. [See below](#authorizer_configuration-block).
* `authorizer_type` - Type of authorizer that controls how consumers access the registry's search and MCP invoke operations. Valid values: `AWS_IAM`, `CUSTOM_JWT`.

### `authorizer_configuration` Block

* `custom_jwt_authorizer` - Configuration for a custom JWT authorizer. [See below](#custom_jwt_authorizer-block).

### `custom_jwt_authorizer` Block

* `allowed_audience` - Audience values accepted during JWT validation.
* `allowed_clients` - Client identifiers accepted during JWT validation.
* `allowed_scopes` - Scopes accepted during JWT validation.
* `custom_claim` - Custom claims for additional JWT validation beyond standard OIDC claims. [See below](#custom_claim-block).
* `discovery_url` - OpenID Connect discovery URL used to retrieve the identity provider's metadata and signing keys.
* `private_endpoint` - Private endpoint used to reach the identity provider's discovery URL over a private network path. [See below](#private_endpoint-block).
* `private_endpoint_override` - Per-domain private endpoint overrides that route specific identity provider domains through distinct private endpoints. [See below](#private_endpoint_override-block).

### `custom_claim` Block

* `authorizing_claim_match_value` - Claim match criteria. [See below](#authorizing_claim_match_value-block).
* `inbound_token_claim_name` - Name of the claim validated in the inbound JWT token.
* `inbound_token_claim_value_type` - Type of the claim value. Valid values: `STRING`, `STRING_ARRAY`.

### `authorizing_claim_match_value` Block

* `claim_match_operator` - Operator used to match claim values. Valid values: `EQUALS`, `CONTAINS`, `CONTAINS_ANY`.
* `claim_match_value` - Value matched against. [See below](#claim_match_value-block).

### `claim_match_value` Block

* `match_value_string` - Single string value to match.
* `match_value_string_list` - Set of string values to match.

### `private_endpoint` Block

Exactly one of the following must be specified:

* `managed_vpc_resource` - Private endpoint backed by a service-managed VPC resource. [See below](#managed_vpc_resource-block).
* `self_managed_lattice_resource` - Private endpoint backed by a self-managed VPC Lattice resource configuration. [See below](#self_managed_lattice_resource-block).

### `managed_vpc_resource` Block

* `endpoint_ip_address_type` - IP address type used by the private endpoint, either `IPV4` or `IPV6`.
* `routing_domain` - Routing domain used to resolve traffic through the private endpoint.
* `security_group_ids` - IDs of the security groups associated with the private endpoint network interfaces.
* `subnet_ids` - IDs of the subnets in which the private endpoint network interfaces are placed.
* `tags` - Tags applied to the service-managed VPC resource.
* `vpc_identifier` - ID of the VPC in which the private endpoint is provisioned.

### `self_managed_lattice_resource` Block

* `resource_configuration_identifier` - Identifier of the VPC Lattice resource configuration, specified as a resource configuration ID or ARN.

### `private_endpoint_override` Block

* `domain` - Domain name to which this private endpoint override applies.
* `private_endpoint` - Private endpoint used to reach the specified domain. [See above](#private_endpoint-block).

### `encryption_configuration` Block

* `kms_key_arn` - ARN of the customer-managed AWS KMS key used to encrypt the registry's content.
