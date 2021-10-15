---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_control_panel"
description: |-
  Provides an AWS Route 53 Recovery Control Config Control Panel
---

# Resource: aws_route53recoverycontrolconfig_control_panel

Provides an AWS Route 53 Recovery Control Config Control Panel.

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_control_panel" "example" {
  name        = "balmorhea"
  cluster_arn = "arn:aws:route53-recovery-control::123456789012:cluster/8d47920e-d789-437d-803a-2dcc4b204393"
}
```

## Argument Reference

The following arguments are required:

* `cluster_arn` - (Required) ARN of the cluster in which this control panel will reside.
* `name` - (Required) Name describing the control panel.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the control panel.
* `default_control_panel` - Whether a control panel is default.
* `routing_control_count` - Number routing controls in a control panel.
* `status` - Status of control panel: `PENDING` when it is being created/updated, `PENDING_DELETION` when it is being deleted, and `DEPLOYED` otherwise.

## Import

Route53 Recovery Control Config Control Panel can be imported via the control panel arn, e.g.,

```
$ terraform import aws_route53recoverycontrolconfig_control_panel.mypanel arn:aws:route53-recovery-control::313517334327:controlpanel/1bfba17df8684f5dab0467b71424f7e8
```
