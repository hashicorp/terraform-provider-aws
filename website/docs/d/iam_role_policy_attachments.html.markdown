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

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `path_prefix` - (Optional) Path prefix for filtering the results.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attached_policies` - List of attached managed policies. See [below](#attached_policies-block).

### `attached_policies` Block

* `policy_arn` - ARN of the attached managed policy.
* `policy_name` - Name of the attached managed policy.
