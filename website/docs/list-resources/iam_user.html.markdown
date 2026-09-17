---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user"
description: |-
  Lists IAM (Identity & Access Management) User resources.
---

# List Resource: aws_iam_user

Lists IAM (Identity & Access Management) User resources.

## Example Usage

### Basic Usage

```terraform
list "aws_iam_user" "example" {
  provider = aws
}
```

### Filter by Path Prefix

This example will return IAM Users with a `path` equal to or beginning with `/example/`.

```terraform
list "aws_iam_user" "example" {
  provider = aws

  config {
    path_prefix = "/example/"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `path_prefix` - (Optional) Limits the returned IAM Users to those within this path.
  If `path_prefix` is not specified, or is `"/"`, returns all IAM Users.
  Must begin and end with a slash (`/`) and contain uppercase or lowercase alphanumeric characters or any of the following: `/`, `,`, `.`, `+`, `@`, `=`, `_`, or `-`.
