---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policies"
description: |-
  Terraform data source for retreiving inline role policies.
---

# Data Source: aws_iam_role_policies

Terraform data source for retreiving inline role policies.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_role_policies" "example" {
  role_name = "test_role"
}
```

## Argument Reference

The following arguments are required:

* `role_name` - (Required) Role name from which we want to retreive the inline policies.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy_names` - List containing names of all inline policies of the specified role.