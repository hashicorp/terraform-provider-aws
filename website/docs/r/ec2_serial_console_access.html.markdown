---
subcategory: "EC2"
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

The following arguments are supported:

* `enabled` - (Optional) Whether or not serial console access is enabled. Valid values are `true` or `false`. Defaults to `true`.

## Attributes Reference

No additional attributes are exported.

## Import

Serial console access state can be imported, e.g.,

```
$ terraform import aws_ec2_serial_console_access.example default
```
