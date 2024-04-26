---
subcategory: "Route 53 Recovery Control Config"
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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the control panel.
* `default_control_panel` - Whether a control panel is default.
* `routing_control_count` - Number routing controls in a control panel.
* `status` - Status of control panel: `PENDING` when it is being created/updated, `PENDING_DELETION` when it is being deleted, and `DEPLOYED` otherwise.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Recovery Control Config Control Panel using the control panel arn. For example:

```terraform
import {
  to = aws_route53recoverycontrolconfig_control_panel.mypanel
  id = "arn:aws:route53-recovery-control::313517334327:controlpanel/1bfba17df8684f5dab0467b71424f7e8"
}
```

Using `terraform import`, import Route53 Recovery Control Config Control Panel using the control panel arn. For example:

```console
% terraform import aws_route53recoverycontrolconfig_control_panel.mypanel arn:aws:route53-recovery-control::313517334327:controlpanel/1bfba17df8684f5dab0467b71424f7e8
```
