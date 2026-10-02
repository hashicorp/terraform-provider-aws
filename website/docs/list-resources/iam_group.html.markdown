---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group"
description: |-
  Lists IAM (Identity & Access Management) Group resources.
---

# List Resource: aws_iam_group

Lists IAM (Identity & Access Management) Group resources.

## Example Usage

### Basic Usage

```terraform
list "aws_iam_group" "example" {
  provider = aws
}
```

### Filter by Path Prefix

This example returns IAM Groups with a `path` equal to or beginning with `/example/`.

```terraform
list "aws_iam_group" "example" {
  provider = aws

  config {
    path_prefix = "/example/"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `path_prefix` - (Optional) Limits the returned IAM Groups to those within this path. If `path_prefix` is not specified, or is `"/"`, returns all IAM Groups. Must begin and end with a slash (`/`) and contain uppercase or lowercase alphanumeric characters or any of the following: `/`, `,`, `.`, `+`, `@`, `=`, `_`, or `-`.
