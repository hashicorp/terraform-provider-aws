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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
