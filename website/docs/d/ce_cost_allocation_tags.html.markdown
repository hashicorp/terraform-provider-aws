---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: ce_cost_allocation_tags"
description: |-
  Provides the available cost allocation tags.
---

# Data Source: ce_cost_allocation_tags

Provides the available cost allocation tags.

## Example Usage

### Basic Usage

```terraform
data "aws_ce_cost_allocation_tags" "tags" {}
```

### Filter by Status and Type

```terraform
data "aws_ce_cost_allocation_tags" "inactive_tags" {
  status = "Inactive"
  type   = "UserDefined"
}
```

### Filter by Tag Keys

```terraform
data "aws_ce_cost_allocation_tags" "tags" {
  tag_keys = ["tag_a", "tag_b"]
}
```

## Argument Reference

This data source supports the following arguments:

* `status` - (Optional) The status of cost allocation tags that you want to return values for.
* `type` - (Optional) The type of cost allocation tags that you want to return values for.
* `tag_keys` - (Optional) Keys of the tag that you want to return values for.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `tags` - Tags that match your request.

### `tags` Attribute Reference

* `last_updated_date` - The last date that the tag was either activated or deactivated.
* `last_used_date` - The last month that the tag was used on an Amazon Web Services resource.
* `status` - The status of the tag.
* `tag_key` - Key of the tag.
* `type` - The type of the tag.
