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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The key for the cost allocation tag.
* `type` - The type of cost allocation tag.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ce_cost_allocation_tag` using the `id`. For example:

```terraform
import {
  to = aws_ce_cost_allocation_tag.example
  id = "key"
}
```

Using `terraform import`, import `aws_ce_cost_allocation_tag` using the `id`. For example:

```console
% terraform import aws_ce_cost_allocation_tag.example key
```
