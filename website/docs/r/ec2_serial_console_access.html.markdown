---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_serial_console_access"
description: |-
  Manages whether serial console access is enabled for your AWS account in the current AWS region.
---

# Resource: aws_ec2_serial_console_access

Provides a resource to manage whether serial console access is enabled for your AWS account in the current AWS region.

~> **NOTE:** Removing this Terraform resource disables serial console access.

## Example Usage

```terraform
resource "aws_ec2_serial_console_access" "example" {
  enabled = true
}
```

## Argument Reference

This resource supports the following arguments:

* `enabled` - (Optional) Whether or not serial console access is enabled. Valid values are `true` or `false`. Defaults to `true`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import serial console access state. For example:

```terraform
import {
  to = aws_ec2_serial_console_access.example
  id = "default"
}
```

Using `terraform import`, import serial console access state. For example:

```console
% terraform import aws_ec2_serial_console_access.example default
```
