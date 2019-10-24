---
layout: "aws"
page_title: "AWS: aws_iam_group"
description: |-
  Get information on a Amazon IAM group
---

# Data Source: aws_iam_group

This data source can be used to fetch information about a specific
IAM group. By using this data source, you can reference IAM group
properties without having to hard code ARNs as input.

## Example Usage

```hcl
data "aws_iam_group" "example" {
  group_name = "an_example_group_name"
}
```

## Argument Reference

* `group_name` - (Required) The friendly IAM group name to match.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) specifying the group.

* `path` - The path to the group.

* `group_id` - The stable and unique string identifying the group.

* `users` - The list of users in the group.
  * `path` - The path to the user.
  * `user_name` - The friendly name identifying the user.
  * `user_id` - The stable and unique string identifying the user.
  * `arn` - The Amazon Resource Name (ARN) that identifies the user.
  * `create_date` - The date and time the user was created.
  * `password_last_used` - The date and time when the user's password was last used to sign in to an AWS website.
  * `permissions_boundary` - The ARN of the policy used to set the permissions boundary for the user.
    * `permissions_boundary_type` - The permissions boundary usage type for an entity.
    * `permissions_boundary_arn` - The ARN of the policy used to set the permissions boundary for the user or role.
  * `tags` - Tags associated to the user.
    * `tags.#.key` - The key name of the tag.
    * `tags.#.value` - The value of the tag.
