---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_roles"
description: |-
  Get information about a set of IAM Roles.
---

# Data Source: aws_iam_roles

Use this data source to get the ARNs and Names of IAM Roles.

## Example Usage

### All roles in an account

```terraform
data "aws_iam_roles" "roles" {}
```

### Roles filtered by name regex

Roles whose role-name contains `project`

```terraform
data "aws_iam_roles" "roles" {
  name_regex = ".*project.*"
}
```

### Roles filtered by path prefix

```terraform
data "aws_iam_roles" "roles" {
  path_prefix = "/custom-path"
}
```

### Roles provisioned by AWS SSO

Roles in the account filtered by path prefix

```terraform
data "aws_iam_roles" "roles" {
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
```

Specific role in the account filtered by name regex and path prefix

```terraform
data "aws_iam_roles" "roles" {
  name_regex  = "AWSReservedSSO_permission_set_name_.*"
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
```

### Role ARNs with paths removed

For services like Amazon EKS that do not permit a path in the role ARN when used in a cluster's configuration map

```terraform
data "aws_iam_roles" "roles" {
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}

output "arns" {
  value = [
    for parts in [for arn in data.aws_iam_roles.roles.arns : split("/", arn)] :
    format("%s/%s", parts[0], element(parts, length(parts) - 1))
  ]
}
```

## Argument Reference

This data source supports the following arguments:

* `name_regex` - (Optional) Regex string to apply to the IAM roles list returned by AWS. This allows more advanced filtering not supported from the AWS API. This filtering is done locally on what AWS returns, and could have a performance impact if the result is large. Combine this with other options to narrow down the list AWS returns.
* `path_prefix` - (Optional) Path prefix for filtering the results. For example, the prefix `/application_abc/component_xyz/` gets all roles whose path starts with `/application_abc/component_xyz/`. If it is not included, it defaults to a slash (`/`), listing all roles. For more details, check out [list-roles in the AWS CLI reference][1].

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the matched IAM roles.
* `names` - Set of Names of the matched IAM roles.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-roles.html
