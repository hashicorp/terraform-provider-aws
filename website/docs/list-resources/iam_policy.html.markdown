---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Lists IAM Policy resources.
---

# List Resource: aws_iam_policy

~> **Note:** The `aws_iam_policy` List Resource is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Lists IAM Policy resources.

Excludes AWS Managed Policies (see "AWS managed policies" in [Policies and permissions in AWS Identity and Access Management documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html#access_policy-types).

## Example Usage

### Basic Usage

```terraform
list "aws_iam_policy" "example" {
  provider = aws
}
```

### Restricting Path

This example will return IAM Policies with a `path` equal to or beginning with `/example/`.

```terraform
list "aws_iam_policy" "example" {
  provider = aws

  config {
    path_prefix = "/example/"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `path_prefix` - (Optional) Limits the returned IAM Policies to those within this path.
  If `path_prefix` is not specified, or is `"/"`, returns all IAM Policies.
  Must begin and end with a slash (`/`) and contain uppercase or lowercase alphanumeric characters or any of the following: `/`, `,`, `.`, `+`, `@`, `=`, `_`, or `-`.
