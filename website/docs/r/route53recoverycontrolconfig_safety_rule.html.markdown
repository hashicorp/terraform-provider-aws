---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_safety_rule"
description: |-
  Provides an AWS Route 53 Recovery Control Config Safety Rule
---

# Resource: aws_route53recoverycontrolconfig_safety_rule

Provides an AWS Route 53 Recovery Control Config Safety Rule

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_safety_rule" "myassertionrule" {
  name        = aws_route53recoverycontrolconfig_assertion_rule
  control_panel_arn = i_belong_to_this_control_panel
  wait_period_ms = 50000
  rule_config = { inverted = false, threshold = 1, type = ATLEAST}
  asserted_controls = [arn1, arn2]
}
```

```terraform
resource "aws_route53recoverycontrolconfig_safety_rule" "mygatingrule" {
  name        = aws_route53recoverycontrolconfig_gating_rule
  control_panel_arn = i_belong_to_this_control_panel
  wait_period_ms = 50000
  rule_config = { inverted = false, threshold = 1, type = ATLEAST}
  gating_controls = [arn1, arn2]
  target_controls = [arn1, arn2]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name describing the safety rule
* `control_panel_arn` - (Required) ARN of the control panel in which this safety rule will reside
* `wait_period_ms` - (Rquired) An evaluation period, in milliseconds (ms), during which any request against the target routing controls will fail
* `rule_config` - (Required) The criteria that you set for specific safety rules that designate how many controls must be enabled as the result of a transaction
* `inverted` - (Required) Logical negation of the rule. If the rule would usually evaluate true, it's evaluated as false, and vice versa.
* `Threshold` - (Required) The value of N, when you specify an ATLEAST rule type. That is, Threshold is the number of controls that must be set when you specify an ATLEAST type
* `type` - (Required) A rule can be one of the following: ATLEAST, AND, or OR
* `asserted_controls` - The routing controls that are part of transactions that are evaluated to determine if a request to change a routing control state is allowed.
* `gating_controls` - The gating controls for the new gating rule. That is, routing controls that are evaluated by the rule configuration that you specify
* `target_controls` - Routing controls that can only be set or unset if the specified RuleConfig evaluates to true for the specified GatingControls.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `status` - Represents status of safety rule. PENDING when its being created/updated, PENDING_DELETION when its being deleted and DEPLOYED otherwise

## Import

Route53 Recovery Control Config Safety Rule can be imported via the safety rule arn, e.g.

```
$ terraform import aws_route53recoverycontrolconfig_safety_rule.myrule myrule
```

## Timeouts

`aws_route53recoverycontrolconfig_safety_rule` has a timeout of 1 minute for creation, updation and deletion