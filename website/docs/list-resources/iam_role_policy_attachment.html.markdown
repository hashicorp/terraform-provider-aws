---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policy_attachment"
description: |-
  Lists IAM Role-Policy Attachment resources.
---

# List Resource: aws_iam_role_policy_attachment

~> **Note:** The `aws_iam_role_policy_attachment` List Resource is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Lists IAM Role-Policy Attachment resources.

Excludes Service-Linked Roles (see "AWS service-linked role" in [IAM Roles Terms and Concepts documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html#id_roles_terms-and-concepts)).

## Example Usage

```terraform
list "aws_iam_role_policy_attachment" "example" {
  provider = aws
}
```

## Argument Reference

This list resource does not support any arguments.
