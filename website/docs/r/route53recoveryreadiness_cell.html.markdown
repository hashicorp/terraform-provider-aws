---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_cell"
description: |-
  Provides an AWS Route 53 Recovery Readiness Cell
---

# Resource: aws_route53recoveryreadiness_cell

Provides an AWS Route 53 Recovery Readiness Cell.

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_cell" "example" {
  cell_name = "us-west-2-failover-cell"
}
```

## Argument Reference

The following arguments are required:

* `cell_name` - (Required) Unique name describing the cell.

The following arguments are optional:

* `cells` - (Optional) List of cell arns to add as nested fault domains within this cell.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the cell
* `parent_readiness_scopes` - List of readiness scopes (recovery groups or cells) that contain this cell.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness cells can be imported via the cell name, e.g.,

```
$ terraform import aws_route53recoveryreadiness_cell.us-west-2-failover-cell us-west-2-failover-cell
```

## Timeouts

`aws_route53recoveryreadiness_cell` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `5m`) Used when deleting the Cell
