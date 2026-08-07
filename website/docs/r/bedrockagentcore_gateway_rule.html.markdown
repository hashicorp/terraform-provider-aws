---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_rule"
description: |-
  Manages an AWS Bedrock AgentCore Gateway Rule.
---

# Resource: aws_bedrockagentcore_gateway_rule

Manages an AWS Bedrock AgentCore Gateway Rule. Rules define conditions and actions that control how requests are routed and processed through a gateway, including principal-based access control, path-based routing, weighted target routing, and configuration bundle overrides. Rules are evaluated in order of `priority` (lower numbers first).

## Example Usage

### Route to a Static Target

```terraform
resource "aws_bedrockagentcore_gateway_rule" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  priority           = 100
  description        = "Route all requests to the primary target"

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.example.name
      }
    }
  }
}
```

### Weighted Route (Canary Traffic Split)

```terraform
resource "aws_bedrockagentcore_gateway_rule" "canary" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  priority           = 100

  action {
    route_to_target {
      weighted_route {
        traffic_split {
          name        = "primary"
          target_name = aws_bedrockagentcore_gateway_target.primary.name
          weight      = 90
        }
        traffic_split {
          name        = "canary"
          target_name = aws_bedrockagentcore_gateway_target.canary.name
          weight      = 10
        }
      }
    }
  }
}
```

### Match on IAM Principals and Paths

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_bedrockagentcore_gateway_rule" "restricted" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  priority           = 50

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.example.name
      }
    }
  }

  condition {
    match_principals {
      any_of {
        iam_principal {
          arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/agentcore-caller-*"
          operator = "StringLike"
        }
      }
    }
  }

  condition {
    match_paths {
      any_of = ["/api/*"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `action` - (Required) One or two [`action`](#action) blocks defining what happens when the rule's conditions match. See [Action](#action) below.
* `gateway_identifier` - (Required, Forces new resource) Identifier of the gateway to attach the rule to.
* `priority` - (Required) Priority of the rule, between 1 and 1000000. Rules are evaluated in ascending order of priority.

The following arguments are optional:

* `condition` - (Optional) Up to two [`condition`](#condition) blocks that must all be satisfied for the rule's actions to apply. See [Condition](#condition) below.
* `description` - (Optional) Description of the rule. Between 1 and 256 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `action` Block

Exactly one of `configuration_bundle` or `route_to_target` must be set on each `action` block.

* `configuration_bundle` - (Optional) Apply a configuration bundle when the rule's conditions match. See [configuration_bundle](#configuration_bundle) below.
* `route_to_target` - (Optional) Route requests to a gateway target when the rule's conditions match. See [route_to_target](#route_to_target) below.

### `action.configuration_bundle` Block

Exactly one of `static_override` or `weighted_override` must be set.

* `static_override` - (Optional) Statically override the configuration bundle used for the matched request.
* `weighted_override` - (Optional) Distribute the request across two configuration bundle versions by weight.

### `action.configuration_bundle.static_override` Block

* `bundle_arn` - (Required) ARN of the configuration bundle to apply.
* `bundle_version` - (Required) Version (UUID) of the configuration bundle to apply.

### `action.configuration_bundle.weighted_override` Block

* `traffic_split` - (Required) Exactly two `traffic_split` blocks describing the two variants.

### `action.configuration_bundle.weighted_override.traffic_split` Block

* `configuration_bundle` - (Required) Reference to the configuration bundle for this variant.
* `description` - (Optional) Description of the variant. Between 1 and 200 characters.
* `metadata` - (Optional) Up to 25 key/value metadata pairs describing this variant.
* `name` - (Required) Name of this variant. Between 1 and 64 characters; alphanumeric with internal hyphens.
* `weight` - (Required) Percentage of traffic sent to this variant, between 1 and 99. Weights across the two entries must sum to 100.

### `action.configuration_bundle.weighted_override.traffic_split.configuration_bundle` Block

* `bundle_arn` - (Required) ARN of the configuration bundle to apply.
* `bundle_version` - (Required) Version (UUID) of the configuration bundle to apply.

### `action.route_to_target` Block

Exactly one of `static_route` or `weighted_route` must be set.

* `static_route` - (Optional) Route all matching requests to a single named gateway target.
* `weighted_route` - (Optional) Distribute requests across two named targets by weight.

### `action.route_to_target.static_route` Block

* `target_name` - (Required) Name of the gateway target.

### `action.route_to_target.weighted_route` Block

* `traffic_split` - (Required) Exactly two `traffic_split` blocks describing the two variants.

### `action.route_to_target.weighted_route.traffic_split` Block

* `description` - (Optional) Description of the variant. Between 1 and 200 characters.
* `metadata` - (Optional) Up to 25 key/value metadata pairs describing this variant.
* `name` - (Required) Name of this variant. Between 1 and 64 characters; alphanumeric with internal hyphens.
* `target_name` - (Required) Name of the gateway target this variant points to.
* `weight` - (Required) Percentage of traffic routed to this variant, between 1 and 99.

### `condition` Block

Exactly one of `match_paths` or `match_principals` must be set.

* `match_paths` - (Optional) Match when the request path matches any of the supplied glob patterns (e.g. `/api/*`).
* `match_principals` - (Optional) Match when the caller's IAM identity matches any of the supplied principal entries.

### `condition.match_paths` Block

* `any_of` - (Required) Between 1 and 10 path patterns. A pattern must be of the form `/<segment>/*` and at most 512 characters.

### `condition.match_principals` Block

* `any_of` - (Required) Between 1 and 100 principal entry blocks.

### `condition.match_principals.any_of` Block

* `iam_principal` - (Required) Match an IAM user, role, or assumed-role ARN. Exactly one `iam_principal` block is required per entry.

### `condition.match_principals.any_of.iam_principal` Block

* `arn` - (Required) IAM principal ARN. Wildcards are allowed with the `StringLike` operator.
* `operator` - (Optional) Match operator, one of `StringEquals` or `StringLike`. Defaults to `StringEquals`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `gateway_arn` - ARN of the gateway that owns the rule.
* `rule_id` - Identifier of the rule.
* `system` - Present when the rule is system-managed. A single-element list with a `managed_by` string.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_rule.example
  identity = {
    gateway_identifier = "example-gateway-abcdef1234"
    rule_id            = "11111111-2222-3333-4444-555555555555"
  }
}

resource "aws_bedrockagentcore_gateway_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `gateway_identifier` (String) Identifier of the gateway.
* `rule_id` (String) Identifier of the rule.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import gateway rules using `gateway_identifier` and `rule_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_rule.example
  id = "example-gateway-abcdef1234,11111111-2222-3333-4444-555555555555"
}
```

Using `terraform import`, import gateway rules using `gateway_identifier` and `rule_id` separated by a comma (`,`). For example:

```console
% terraform import aws_bedrockagentcore_gateway_rule.example example-gateway-abcdef1234,11111111-2222-3333-4444-555555555555
```
