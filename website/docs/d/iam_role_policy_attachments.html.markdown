---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policy_attachments"
description: |-
  Provides details about AWS IAM Role Policy Attachments.
---

# Data Source: aws_iam_role_policy_attachments

Provides details about the managed policies attached to an AWS IAM Role.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_role_policy_attachments" "example" {
  role_name = "example-role"
}
```

## Argument Reference

The following arguments are required:

* `role_name` - (Required) Name of the IAM role.

The following arguments are optional:

* `path_prefix` - (Optional) Path prefix for filtering the results.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attached_policies` - List of attached managed policies. See [below](#attached_policies-attribute-reference).

### `attached_policies` Attribute Reference

* `policy_arn` - ARN of the attached managed policy.
* `policy_name` - Name of the attached managed policy.
