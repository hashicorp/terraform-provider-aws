---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Get information on a Amazon IAM policy
---

# aws_iam_policy

This data source can be used to fetch information about a specific
IAM policy.

## Example Usage

```terraform
data "aws_iam_policy" "example" {
  arn = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
}
```
Using IAM policy `name` as input,

```hcl
data "aws_iam_policy" "example" {
  name = "test_policy"
}
```

## Argument Reference

* `arn` - (Optional) The ARN of the IAM policy.
* `name` - (Optional) The name of the IAM policy. You must use either ARN or name to fetch information about IAM policy.
* `path_prefix` - (Optional) The path to the IAM policy. This parameter allows a string of characters consisting of either a forward slash (/) by itself or a string that must begin and end with forward slashes. E.G. `/service-role/`

## Attributes Reference

* `name` - The name of the IAM policy.
* `arn` - The Amazon Resource Name (ARN) specifying the policy.
* `path` - The path to the policy.
* `description` - The description of the policy.
* `policy` - The policy document of the policy.
* `policy_id` - The policy's ID.
* `tags` - Key-value mapping of tags for the IAM Policy

