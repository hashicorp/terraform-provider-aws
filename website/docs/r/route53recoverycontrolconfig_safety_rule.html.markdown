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
resource "aws_route53recoverycontrolconfig_safety_rule" "example" {
  asserted_controls = [aws_route53recoverycontrolconfig_routing_control.example.arn]
  control_panel_arn = "arn:aws:route53-recovery-control::313517334327:controlpanel/abd5fbfc052d4844a082dbf400f61da8"
  name              = "daisyguttridge"
  wait_period_ms    = 5000

  rule_config {
    inverted  = false
    threshold = 1
    type      = "ATLEAST"
  }
}
```

```terraform
resource "aws_route53recoverycontrolconfig_safety_rule" "example" {
  name              = "i_o"
  control_panel_arn = "arn:aws:route53-recovery-control::313517334327:controlpanel/abd5fbfc052d4844a082dbf400f61da8"
  wait_period_ms    = 5000
  gating_controls   = [aws_route53recoverycontrolconfig_routing_control.example.arn]
  target_controls   = [aws_route53recoverycontrolconfig_routing_control.example.arn]

  rule_config {
    inverted  = false
    threshold = 1
    type      = "ATLEAST"
  }
}
```

## Argument Reference

The following arguments are supported:

* `control_panel_arn` - (Required) ARN of the control panel in which this safety rule will reside.
* `name` - (Required) Name describing the safety rule.
* `rule_config` - (Required) Configuration block for safety rule criteria. See below.
* `wait_period_ms` - (Required) Evaluation period, in milliseconds (ms), during which any request against the target routing controls will fail.

The following arguments are optional:

* `asserted_controls` - (Optional) Routing controls that are part of transactions that are evaluated to determine if a request to change a routing control state is allowed.
* `gating_controls` - (Optional) Gating controls for the new gating rule. That is, routing controls that are evaluated by the rule configuration that you specify.
* `target_controls` - (Optional) Routing controls that can only be set or unset if the specified `rule_config` evaluates to true for the specified `gating_controls`.

### rule_config

* `inverted` - (Required) Logical negation of the rule.
* `threshold` - (Required) Number of controls that must be set when you specify an `ATLEAST` type rule.
* `type` - (Required) Rule type. Valid values are `ATLEAST`, `AND`, and `OR`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the safety rule.
* `status` - Status of the safety rule. `PENDING` when it is being created/updated, `PENDING_DELETION` when it is being deleted, and `DEPLOYED` otherwise.

## Import

Route53 Recovery Control Config Safety Rule can be imported via the safety rule ARN, e.g.,

```
$ terraform import aws_route53recoverycontrolconfig_safety_rule.myrule arn:aws:route53-recovery-control::313517334327:controlpanel/1bfba17df8684f5dab0467b71424f7e8/safetyrule/3bacc77003364c0f
```
