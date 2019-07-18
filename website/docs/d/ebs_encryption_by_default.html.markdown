---
layout: "aws"
page_title: "AWS: aws_ebs_encryption_by_default"
sidebar_current: "docs-aws-ebs-encryption-by-default"
description: |-
  Checks whether default EBS encryption is enabled for your AWS account in the current AWS region.
---

# Data Source: aws_ebs_encryption_by_default

Provides a way to check whether default EBS encryption is enabled for your AWS account in the current AWS region.

## Example Usage

```hcl
data "aws_ebs_encryption_by_default" "current" { }
```

## Attributes Reference

The following attributes are exported:

* `enabled` - Whether or not default EBS encryption is enabled. Returns as `true` or `false`.