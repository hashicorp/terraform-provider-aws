---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_user_hierarchy_group"
description: |-
  Provides details about a specific Amazon Connect User Hierarchy Group
---

# Resource: aws_connect_user_hierarchy_group

Provides an Amazon Connect User Hierarchy Group resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

~> **NOTE:** The User Hierarchy Structure must be created before creating a User Hierarchy Group.

## Example Usage

### Basic

```terraform
resource "aws_connect_user_hierarchy_group" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "example"

  tags = {
    "Name" = "Example User Hierarchy Group"
  }
}
```

### With a parent group

```terraform
resource "aws_connect_user_hierarchy_group" "parent" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "parent"

  tags = {
    "Name" = "Example User Hierarchy Group Parent"
  }
}

resource "aws_connect_user_hierarchy_group" "child" {
  instance_id     = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name            = "child"
  parent_group_id = aws_connect_user_hierarchy_group.parent.hierarchy_group_id

  tags = {
    "Name" = "Example User Hierarchy Group Child"
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `name` - (Required) The name of the user hierarchy group. Must not be more than 100 characters.
* `parent_group_id` - (Optional) The identifier for the parent hierarchy group. The user hierarchy is created at level one if the parent group ID is null.
* `tags` - (Optional) Tags to apply to the hierarchy group. If configured with a provider
[`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the hierarchy group.
* `hierarchy_group_id` - The identifier for the hierarchy group.
* `hierarchy_path` - A block that contains information about the levels in the hierarchy group. The `hierarchy_path` block is documented below.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the hierarchy group
separated by a colon (`:`).
* `level_id` - The identifier of the level in the hierarchy group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

A `hierarchy_path` block supports the following attributes:

* `level_one` - A block that defines the details of level one. The level block is documented below.
* `level_two` - A block that defines the details of level two. The level block is documented below.
* `level_three` - A block that defines the details of level three. The level block is documented below.
* `level_four` - A block that defines the details of level four. The level block is documented below.
* `level_five` - A block that defines the details of level five. The level block is documented below.

A level block supports the following attributes:

* `arn` -  The Amazon Resource Name (ARN) of the hierarchy group.
* `id` -  The identifier of the hierarchy group.
* `name` - The name of the hierarchy group.

## Import

Amazon Connect User Hierarchy Groups can be imported using the `instance_id` and `hierarchy_group_id` separated by a colon (`:`), e.g.,

```
$ terraform import aws_connect_user_hierarchy_group.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5
```
