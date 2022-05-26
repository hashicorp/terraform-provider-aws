---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_user_hierarchy_structure"
description: |-
  Provides details about a specific Amazon Connect User Hierarchy Structure
---

# Data Source: aws_connect_user_hierarchy_structure

Provides details about a specific Amazon Connect User Hierarchy Structure

## Example Usage

```hcl
data "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance

## Attributes Reference

In addition to all of the argument above, the following attributes are exported:

* `hierarchy_structure` - A block that defines the hierarchy structure's levels. The `hierarchy_structure` block is documented below.

A `hierarchy_structure` block supports the following attributes:

* `level_one` - A block that defines the details of level one. The level block is documented below.
* `level_two` - A block that defines the details of level two. The level block is documented below.
* `level_three` - A block that defines the details of level three. The level block is documented below.
* `level_four` - A block that defines the details of level four. The level block is documented below.
* `level_five` - A block that defines the details of level five. The level block is documented below.

Each level block supports the following attributes:

* `arn` -  The Amazon Resource Name (ARN) of the hierarchy level.
* `id` -  The identifier of the hierarchy level.
* `name` - The name of the user hierarchy level. Must not be more than 50 characters.
