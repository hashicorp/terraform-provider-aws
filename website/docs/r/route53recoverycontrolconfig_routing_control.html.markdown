---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_routing_control"
description: |-
  Provides an AWS Route 53 Recovery Control Config Routing Control
---

# Resource: aws_route53recoverycontrolconfig_routing_control

Provides an AWS Route 53 Recovery Control Config Routing Control.

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_routing_control" "example" {
  name        = "tinlicker"
  cluster_arn = "arn:aws:route53-recovery-control::881188118811:cluster/8d47920e-d789-437d-803a-2dcc4b204393"
}
```

```terraform
resource "aws_route53recoverycontrolconfig_routing_control" "example" {
  name              = "thomasoliver"
  cluster_arn       = "arn:aws:route53-recovery-control::881188118811:cluster/8d47920e-d789-437d-803a-2dcc4b204393"
  control_panel_arn = "arn:aws:route53-recovery-control::428113431245:controlpanel/abd5fbfc052d4844a082dbf400f61da8"
}
```

## Argument Reference

The following arguments are required:

* `cluster_arn` - (Required) ARN of the cluster in which this routing control will reside.
* `name` - (Required) The name describing the routing control.

The following arguments are optional:

* `control_panel_arn` - (Optional) ARN of the control panel in which this routing control will reside.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the routing control.
* `status` - Status of routing control. `PENDING` when it is being created/updated, `PENDING_DELETION` when it is being deleted, and `DEPLOYED` otherwise.

## Import

Route53 Recovery Control Config Routing Control can be imported via the routing control arn, e.g.,

```
$ terraform import aws_route53recoverycontrolconfig_routing_control.mycontrol arn:aws:route53-recovery-control::313517334327:controlpanel/abd5fbfc052d4844a082dbf400f61da8/routingcontrol/d5d90e587870494b
```
