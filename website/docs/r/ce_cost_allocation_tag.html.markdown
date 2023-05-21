---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_cost_allocation_tag"
description: |-
  Provides a CE Cost Allocation Tag
---

# Resource: aws_ce_cost_allocation_tag

Provides a CE Cost Allocation Tag.

## Example Usage

```terraform
resource "aws_ce_cost_allocation_tag" "example" {
  tag_key = "example"
  status  = "Active"
}
```

## Argument Reference

The following arguments are required:

* `tag_key` - (Required) The key for the cost allocation tag.
* `status` - (Required) The status of a cost allocation tag. Valid values are `Active` and `Inactive`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The key for the cost allocation tag.
* `type` - The type of cost allocation tag.

## Import

`aws_ce_cost_allocation_tag` can be imported using the `id`, e.g.

```
$ terraform import aws_ce_cost_allocation_tag.example key
```
