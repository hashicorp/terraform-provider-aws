---
subcategory: "IAM (Identity & Access Management)"
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

```terraform
data "aws_iam_group" "example" {
  group_name = "an_example_group_name"
}
```

## Argument Reference

* `group_name` - (Required) Friendly IAM group name to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Group ARN.
* `group_id` - Stable and unique string identifying the group.
* `id` - Stable and unique string identifying the group.
* `path` - Path to the group.
* `users` - List of objects containing group member information. See below.

### `users`

* `arn` - User ARN.
* `path` - Path to the IAM user.
* `user_id` - Stable and unique string identifying the IAM user.
* `user_name` - Name of the IAM user.
