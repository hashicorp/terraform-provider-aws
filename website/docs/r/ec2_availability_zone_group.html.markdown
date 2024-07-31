---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_availability_zone_group"
description: |-
  Manages an EC2 Availability Zone Group.
---

# Resource: aws_ec2_availability_zone_group

Manages an EC2 Availability Zone Group, such as updating its opt-in status.

~> **NOTE:** This is an advanced Terraform resource. Terraform will automatically assume management of the EC2 Availability Zone Group without import and perform no actions on removal from configuration.

## Example Usage

```terraform
resource "aws_ec2_availability_zone_group" "example" {
  group_name    = "us-west-2-lax-1"
  opt_in_status = "opted-in"
}
```

## Argument Reference

The following arguments are required:

* `group_name` - (Required) Name of the Availability Zone Group.
* `opt_in_status` - (Required) Indicates whether to enable or disable Availability Zone Group. Valid values: `opted-in` or `not-opted-in`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the Availability Zone Group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Availability Zone Groups using the group name. For example:

```terraform
import {
  to = aws_ec2_availability_zone_group.example
  id = "us-west-2-lax-1"
}
```

Using `terraform import`, import EC2 Availability Zone Groups using the group name. For example:

```console
% terraform import aws_ec2_availability_zone_group.example us-west-2-lax-1
```
