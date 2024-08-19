---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_sets"
description: |-
  Terraform data source returning the ARN of all AWS SSO Admin Permission Sets.
---

# Data Source: aws_ssoadmin_permission_sets

Terraform data source returning the ARN of all AWS SSO Admin Permission Sets.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_ssoadmin_permission_sets" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the SSO Instance associated with the permission set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of string contain the ARN of all Permission Sets.
