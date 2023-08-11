---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_users"
description: |-
  Get information about a set of IAM Users.
---

# Data Source: aws_iam_users

Use this data source to get the ARNs and Names of IAM Users.

## Example Usage

### All users in an account

```terraform
data "aws_iam_users" "users" {}
```

### Users filtered by name regex

Users whose username contains `abc`

```terraform
data "aws_iam_users" "users" {
  name_regex = ".*abc.*"
}
```

### Users filtered by path prefix

```terraform
data "aws_iam_users" "users" {
  path_prefix = "/custom-path"
}
```

## Argument Reference

This data source supports the following arguments:

* `name_regex` - (Optional) Regex string to apply to the IAM users list returned by AWS. This allows more advanced filtering not supported from the AWS API. This filtering is done locally on what AWS returns, and could have a performance impact if the result is large. Combine this with other options to narrow down the list AWS returns.
* `path_prefix` - (Optional) Path prefix for filtering the results. For example, the prefix `/division_abc/subdivision_xyz/` gets all users whose path starts with `/division_abc/subdivision_xyz/`. If it is not included, it defaults to a slash (`/`), listing all users. For more details, check out [list-users in the AWS CLI reference][1].

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the matched IAM users.
* `names` - Set of Names of the matched IAM users.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-users.html
