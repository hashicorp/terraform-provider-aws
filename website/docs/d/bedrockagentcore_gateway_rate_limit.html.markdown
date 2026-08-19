---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_rate_limit"
description: |-
  Terraform data source for managing an AWS Bedrock AgentCore Gateway Rate Limit.
---

# Data Source: aws_bedrockagentcore_gateway_rate_limit

Terraform data source for managing an AWS Bedrock AgentCore Gateway Rate Limit.

Looks up a single rate limit on a gateway by its identifier.

~> **Note:** `entries` is a list here, whereas the corresponding resource uses repeatable `entries` blocks. Fully computed blocks are not supported by Terraform protocol V6, which the AWS provider will adopt in a future major version, so read-only nested objects are exposed as list attributes. This data source also exports `status`, `created_at` and `updated_at`, which the resource intentionally omits.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrockagentcore_gateway_rate_limit" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  rate_limit_id      = "per-caller"
}
```

### Checking a Rate Limit Is Active

```terraform
data "aws_bedrockagentcore_gateway_rate_limit" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  rate_limit_id      = "per-caller"
}

output "enforced" {
  value = data.aws_bedrockagentcore_gateway_rate_limit.example.status == "ACTIVE"
}
```

## Argument Reference

The following arguments are required:

* `gateway_identifier` - (Required) Identifier of the gateway the rate limit belongs to.
* `rate_limit_id` - (Required) Identifier of the rate limit.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the rate limit was created, in RFC3339 format.
* `description` - Description of the rate limit's purpose.
* `dimension_keys` - Ordered list of dimension key names determining how traffic is grouped into buckets.
* `entries` - Entries mapping dimension values to rate configurations. See [`entries` Block](#entries-block) below.
* `status` - Current status of the rate limit. One of `CREATING`, `ACTIVE`, `UPDATING`, or `DELETING`.
* `updated_at` - Timestamp when the rate limit was last updated, in RFC3339 format.

### `entries` Block

* `connections` - Concurrent connection limit. See [`connections` Block](#connections-block) below.
* `dimensions` - Map of dimension name to dimension value. A value of `*` is a wildcard matching any value.
* `requests` - Request rate limit. See [`requests` Block](#requests-block) below.
* `tokens` - Token rate limit. See [`tokens` Block](#tokens-block) below.

### `connections` Block

* `period` - Time period. Always `second` for connection limits.
* `rate` - Allowed rate. A value of `0` blocks all matching traffic.

### `requests` Block

* `period` - Time period, either `second` or `minute`.
* `rate` - Allowed rate. A value of `0` blocks all matching traffic.

### `tokens` Block

* `period` - Time period. Always `minute` for token limits.
* `rate` - Allowed rate. A value of `0` blocks all matching traffic.
