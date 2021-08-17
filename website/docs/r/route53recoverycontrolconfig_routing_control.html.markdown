---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_routing_control"
description: |-
  Provides an AWS Route 53 Recovery Control Config Routing Control
---

# Resource: aws_route53recoverycontrolconfig_routing_control

Provides an AWS Route 53 Recovery Control Config Routing Control

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_routing_control" "mycontrol" {
  name        = aws_route53recoverycontrolconfig_routing_control
  cluster_arn = i_belong_to_this_cluster
}
```

```terraform
resource "aws_route53recoverycontrolconfig_routing_control" "mycontrol" {
  name              = aws_route53recoverycontrolconfig_routing_control
  cluster_arn       = i_belong_to_this_cluster
  control_panel_arn = i_have_to_belong_this_control_panel_in_cluster
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name describing the routing control
* `cluster_arn` - (Required) ARN of the cluster in which this routing control will reside
* `control_panel_arn` - (Optional) ARN of the control panel in which this routing control will reside

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `routing_control_arn` - The ARN of the routing control.
* `status` - Represents status of routing control. PENDING when its being created/updated, PENDING_DELETION when its being deleted and DEPLOYED otherwise

## Import

Route53 Recovery Control Config Routing Control can be imported via the routing control arn, e.g.

```
$ terraform import aws_route53recoverycontrolconfig_routing_control.mycontrol mycontrol
```

## Timeouts

`aws_route53recoverycontrolconfig_routing_control` has a timeout of 1 minute for creation, updation and deletion