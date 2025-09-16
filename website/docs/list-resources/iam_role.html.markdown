---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role"
description: |-
  Lists IAM Role resources.
---

# Resource: aws_iam_role

Lists IAM Role resources.

Excludes Service-Linked Roles (see "AWS service-linked role" in [IAM Roles Terms and Concepts documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html#id_roles_terms-and-concepts)).

## Example Usage

```terraform
list "aws_iam_role" "example" {
  provider = aws
}
```

## Argument Reference

This list resource does not support any arguments.
