---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_control_panel"
description: |-
  Provides an AWS Route 53 Recovery Control Config Control Panel
---

# Resource: aws_route53recoverycontrolconfig_control_panel

Provides an AWS Route 53 Recovery Control Config Control Panel

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_control_panel" "mypanel" {
  name = "mypanel"
  cluster_arn = "i_belong_to_this_cluster"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name describing the control panel
* `cluster_arn` - (Required) ARN of the cluster in which this control panel will reside

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `control_panel_arn` - The ARN of the control panel
* `cluster_arn` - ARN of the cluster that the control panel belongs to
* `default_control_panel` - This is true if a control panel is default, false otherwise
* `routing_contol_count` - The number routing controls in a control panel
* `status` - Represents status of control panel. PENDING when its being created/updated, PENDING_DELETION when its being deleted and DEPLOYED otherwise

## Import

Route53 Recovery Control Config Control Panel can be imported via the control panel arn, e.g.

```
$ terraform import aws_route53recoverycontrolconfig_control_panel.mypanel mypanel
```

## Timeouts

`aws_route53recoverycontrolconfig_control_panel` has a timeout of 1 minute for creation, updation and deletion