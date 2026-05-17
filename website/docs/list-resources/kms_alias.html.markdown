---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_alias"
description: |-
  Lists KMS aliases.
---

# List Resource: aws_kms_alias

Lists KMS aliases.

## Example Usage

```terraform
list "aws_kms_alias" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
