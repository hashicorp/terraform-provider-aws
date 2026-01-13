---
subcategory: "EC2 (Elastic Compute Cloud)"
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

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `enabled` - Whether or not serial console access is enabled. Returns as `true` or `false`.
* `id` - Region of serial console access.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
