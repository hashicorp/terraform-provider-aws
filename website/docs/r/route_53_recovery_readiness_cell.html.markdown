---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route_53_recovery_readiness_cell"
description: |-
  Provides an AWS Route 53 Recovery Readiness Cell
---

# Resource: aws_route_53_recovery_readiness_cell

Provides an AWS Route 53 Recovery Readiness Cell

## Example Usage

```terraform
resource "aws_route_53_recovery_readiness_cell" "us-west-2-failover-cell" {
  cell_name  = "us-west-2-failover-cell"
}
```

## Argument Reference

The following arguments are supported:

* `cell` - (Required) A unique identifier describing the channel
* `cells` - (Optional) A list of cell arns to add as nested fault domains within this cell

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cell_arn` - The ARN of the cell
* `parent_readiness_scopes` - A list of readiness scopes (recovery groups or cells) that contain this cell


## Import

Route53 Recovery Readiness cells can be imported via the cell name, e.g.

```
$ terraform import aws_route_53_recovery_readiness_cell.us-west-2-failover-cell us-west-2-failover-cell
```
