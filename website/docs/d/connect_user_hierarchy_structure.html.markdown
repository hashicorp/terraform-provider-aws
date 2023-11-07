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

This data source supports the following arguments:

* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `hierarchy_structure` - Block that defines the hierarchy structure's levels. The `hierarchy_structure` block is documented below.

A `hierarchy_structure` block supports the following attributes:

* `level_one` - Details of level one. See below.
* `level_two` - Details of level two. See below.
* `level_three` - Details of level three. See below.
* `level_four` - Details of level four. See below.
* `level_five` - Details of level five. See below.

Each level block supports the following attributes:

* `arn` -  ARN of the hierarchy level.
* `id` -  The identifier of the hierarchy level.
* `name` - Name of the user hierarchy level. Must not be more than 50 characters.
