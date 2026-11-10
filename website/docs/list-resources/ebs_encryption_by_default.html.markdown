---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ebs_encryption_by_default"
description: |-
  Lists EC2 (Elastic Compute Cloud) Ebs Encryption By Default resources.
---

# List Resource: aws_ebs_encryption_by_default

Lists EC2 (Elastic Compute Cloud) Ebs Encryption By Default resources.

## Example Usage

```terraform
list "aws_ebs_encryption_by_default" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
