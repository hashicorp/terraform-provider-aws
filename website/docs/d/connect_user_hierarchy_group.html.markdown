---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_user_hierarchy_group"
description: |-
  Provides details about a specific Amazon Connect User Hierarchy Group.
---

# Data Source: aws_connect_user_hierarchy_group

Provides details about a specific Amazon Connect User Hierarchy Group.

## Example Usage

By `name`

```hcl
data "aws_connect_user_hierarchy_group" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `hierarchy_group_id`

```hcl
data "aws_connect_user_hierarchy_group" "example" {
  instance_id        = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  hierarchy_group_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `hierarchy_group_id` is required.

The following arguments are supported:

* `hierarchy_group_id` - (Optional) Returns information on a specific hierarchy group by hierarchy group id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific hierarchy group by name

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - ARN of the hierarchy group.
* `hierarchy_path` - Block that contains information about the levels in the hierarchy group. The `hierarchy_path` block is documented below.
* `level_id` - Identifier of the level in the hierarchy group.
* `id` - Identifier of the hosting Amazon Connect Instance and identifier of the hierarchy group separated by a colon (`:`).
* `tags` - Map of tags to assign to the hierarchy group.

A `hierarchy_path` block supports the following attributes:

* `level_one` - Block that defines the details of level one. The level block is documented below.
* `level_two` - Block that defines the details of level two. The level block is documented below.
* `level_three` - Block that defines the details of level three. The level block is documented below.
* `level_four` - Block that defines the details of level four. The level block is documented below.
* `level_five` - Block that defines the details of level five. The level block is documented below.

A level block supports the following attributes:

* `arn` -  The ARN of the hierarchy group.
* `id` -  The identifier of the hierarchy group.
* `name` - Name of the hierarchy group.
