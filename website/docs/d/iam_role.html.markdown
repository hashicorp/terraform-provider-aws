---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role"
description: |-
  Get information on a Amazon IAM role
---

# Data Source: aws_iam_role

This data source can be used to fetch information about a specific
IAM role. By using this data source, you can reference IAM role
properties without having to hard code ARNs as input.

## Example Usage

```terraform
data "aws_iam_role" "example" {
  name = "an_example_role_name"
}
```

## Argument Reference

* `name` - (Required) Friendly IAM role name to match.

## Attributes Reference

* `id` - Friendly IAM role name to match.
* `arn` - ARN of the role.
* `assume_role_policy` - Policy document associated with the role.
* `create_date` - Creation date of the role in RFC 3339 format.
* `description` - Description for the role.
* `max_session_duration` - Maximum session duration.
* `path` - Path to the role.
* `permissions_boundary` - The ARN of the policy that is used to set the permissions boundary for the role.
* `unique_id` - Stable and unique string identifying the role.
* `tags` - Tags attached to the role.
