---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Get information on a Amazon IAM policy
---

# Data Source: aws_iam_policy

This data source can be used to fetch information about a specific
IAM policy.

## Example Usage

### By ARN

```terraform
data "aws_iam_policy" "example" {
  arn = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
}
```

### By Name

```terraform
data "aws_iam_policy" "example" {
  name = "test_policy"
}
```

## Argument Reference

* `arn` - (Optional) ARN of the IAM policy.
  Conflicts with `name` and `path_prefix`.
* `name` - (Optional) Name of the IAM policy.
  Conflicts with `arn`.
* `path_prefix` - (Optional) Prefix of the path to the IAM policy.
  Defaults to a slash (`/`).
  Conflicts with `arn`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `attachment_count` - Number of entities (users, groups, and roles) that the policy is attached to.
* `path` - Path to the policy.
* `description` - Description of the policy.
* `policy` - Policy document of the policy.
* `policy_id` - Policy's ID.
* `tags` - Key-value mapping of tags for the IAM Policy.
