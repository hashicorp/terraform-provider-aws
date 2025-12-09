---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_key"
description: |-
  Lists KMS keys.
---

# List Resource: aws_kms_key

Lists KMS keys.

## Example Usage

```terraform
list "aws_kms_key" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
