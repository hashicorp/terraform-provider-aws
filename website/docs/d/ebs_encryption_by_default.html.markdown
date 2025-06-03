---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_encryption_by_default"
description: |-
  Checks whether default EBS encryption is enabled for your AWS account in the current AWS region.
---

# Data Source: aws_ebs_encryption_by_default

Provides a way to check whether default EBS encryption is enabled for your AWS account in the current AWS region.

## Example Usage

```terraform
data "aws_ebs_encryption_by_default" "current" {}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `enabled` - Whether or not default EBS encryption is enabled. Returns as `true` or `false`.
* `id` - Region of default EBS encryption.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
