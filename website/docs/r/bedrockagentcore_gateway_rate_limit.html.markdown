---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_rate_limit"
description: |-
  Manages an AWS Bedrock AgentCore Gateway Rate Limit.
---

# Resource: aws_bedrockagentcore_gateway_rate_limit

Manages an AWS Bedrock AgentCore Gateway Rate Limit. Rate limits throttle traffic through a gateway, grouping requests into buckets by one or more dimension keys and setting an allowed rate for each bucket.

A gateway can have up to 50 rate limits, each identified by a unique, immutable ordered list of `dimension_keys`. Create several resources against the same gateway to limit on different dimensions. Within an entry, a pinned dimension value creates a single bucket shared by all matching traffic, while a `*` wildcard creates a separate bucket for each distinct value observed.

Three separate rules govern how limits combine. All rate limits on a gateway must pass for a request to proceed. Within a single rate limit, the most specific matching entry wins and only that one applies. The effective rate is the lower of the configured rate and the AWS service-managed limit. Changes propagate to the data plane within 30 seconds.

~> **Note:** Rate limits use fail-open behavior. If the rate limit service is unavailable or a dimension cannot be resolved from the request, the gateway allows the request through. Do not rely on rate limits as a security boundary.

## Example Usage

### Per-Caller Limit

Each distinct JWT subject gets its own bucket, rather than all callers sharing one.

```terraform
resource "aws_bedrockagentcore_gateway_rate_limit" "per_caller" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  rate_limit_id      = "per-caller"
  description        = "One bucket per JWT subject"

  dimension_keys = ["$.context.jwt.sub"]

  entries {
    dimensions = {
      "$.context.jwt.sub" = "*"
    }

    requests {
      rate   = 100
      period = "second"
    }

    tokens {
      rate   = 50000
      period = "minute"
    }
  }
}
```

### Multiple Dimensions with Wildcards

`dimension_keys` is ordered, and `*` is only valid in trailing positions. The gateway matches the most specific entry first.

```terraform
resource "aws_bedrockagentcore_gateway_rate_limit" "per_target_tool" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  dimension_keys = ["targetName", "toolName"]

  # Most specific: both positions pinned.
  entries {
    dimensions = {
      targetName = "search-target"
      toolName   = "readData"
    }

    requests {
      rate   = 250
      period = "second"
    }
  }

  # Every other tool on this target.
  entries {
    dimensions = {
      targetName = "search-target"
      toolName   = "*"
    }

    requests {
      rate   = 50
      period = "second"
    }
  }

  # Catch-all, matched only when nothing above does.
  entries {
    dimensions = {
      targetName = "*"
      toolName   = "*"
    }

    requests {
      rate   = 10
      period = "second"
    }

    connections {
      rate   = 5
      period = "second"
    }
  }
}
```

### Blocking Traffic

A `rate` of `0` blocks all matching traffic.

```terraform
resource "aws_bedrockagentcore_gateway_rate_limit" "blocked" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  dimension_keys = ["$.context.iam.principal"]

  entries {
    dimensions = {
      "$.context.iam.principal" = "arn:aws:iam::123456789012:role/blocked"
    }

    requests {
      rate   = 0
      period = "second"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `dimension_keys` - (Required) Ordered list of 1 to 10 dimension key names determining how traffic is grouped into buckets. Must be unique per gateway. Valid values are `targetName`, `toolName`, `qualifiedModelId`, `$.context.iam.principal`, `$.context.iam.sourceIdentity`, and `$.context.jwt.<claim>` where `<claim>` is any JWT claim name. Order is significant: it determines which positions may hold the `*` wildcard. Changing this forces a new resource.
* `entries` - (Required) Entries mapping dimension values to rate configurations. Between 1 and 1,000. See [`entries` Block](#entries-block) below.
* `gateway_identifier` - (Required) Identifier of the gateway the rate limit applies to. Changing this forces a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the rate limit's purpose. Up to 512 characters. Removing this argument clears the description.
* `rate_limit_id` - (Optional) Identifier for the rate limit, 2 to 64 characters, beginning and ending with an alphanumeric character. If omitted, one is generated. Changing this forces a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `entries` Block

The `entries` block supports the following. At least one of `connections`, `requests`, or `tokens` must be set on each entry.

* `connections` - (Optional) Concurrent connection limit. `period` must be `second`. See [`connections` Block](#connections-block) below.
* `dimensions` - (Required) Map of dimension name to dimension value. Keys must match `dimension_keys` exactly. A value of `*` is a wildcard matching any value, and may only appear in trailing positions relative to the `dimension_keys` ordering. Values are limited to 256 characters.
* `requests` - (Optional) Request rate limit. See [`requests` Block](#requests-block) below.
* `tokens` - (Optional) Token rate limit. `period` must be `minute`. See [`tokens` Block](#tokens-block) below.

### `connections` Block

The `connections` block supports the following:

* `period` - (Required) Time period. Must be `second`.
* `rate` - (Required) Allowed rate, between `0` and `10000000`. A value of `0` blocks all matching traffic.

### `requests` Block

The `requests` block supports the following:

* `period` - (Required) Time period. Valid values are `second` and `minute`.
* `rate` - (Required) Allowed rate, between `0` and `10000000`. A value of `0` blocks all matching traffic.

### `tokens` Block

The `tokens` block supports the following:

* `period` - (Required) Time period. Must be `minute`.
* `rate` - (Required) Allowed rate, between `0` and `10000000`. A value of `0` blocks all matching traffic.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_rate_limit.example
  identity = {
    gateway_identifier = "example-gateway-abc1234567"
    rate_limit_id      = "per-caller"
  }
}

resource "aws_bedrockagentcore_gateway_rate_limit" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `gateway_identifier` (String) Identifier of the gateway.
* `rate_limit_id` (String) Identifier of the rate limit.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Gateway Rate Limits using the `gateway_identifier` and `rate_limit_id` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_rate_limit.example
  id = "example-gateway-abc1234567,per-caller"
}
```

Using `terraform import`, import Bedrock AgentCore Gateway Rate Limits using the `gateway_identifier` and `rate_limit_id` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_gateway_rate_limit.example example-gateway-abc1234567,per-caller
```
