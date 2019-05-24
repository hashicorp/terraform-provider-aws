---
layout: "aws"
page_title: "AWS: aws_ebs_encryption_by_default"
sidebar_current: "docs-aws-ebs-encryption-by-default"
description: |-
  Manages whether default EBS encryption is enabled for your AWS account in the current AWS region.
---

# Resource: aws_ebs_encryption_by_default

Provides a resource to manage whether default EBS encryption is enabled for your AWS account in the current AWS region.

## Example Usage

```hcl
resource "aws_ebs_encryption_by_default" "example" {
  enabled = true
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Required) Whether or not default EBS encryption is enabled. Valid values are `true` or `false`.
