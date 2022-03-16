---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_serial_console_access"
description: |-
  Checks whether serial console access is enabled for your AWS account in the current AWS region.
---

# Data Source: aws_ec2_serial_console_access

Provides a way to check whether serial console access is enabled for your AWS account in the current AWS region.

## Example Usage

```terraform
data "aws_ec2_serial_console_access" "current" {}
```

## Attributes Reference

The following attributes are exported:

* `enabled` - Whether or not serial console access is enabled. Returns as `true` or `false`.
* `id` - Region of serial console access.
