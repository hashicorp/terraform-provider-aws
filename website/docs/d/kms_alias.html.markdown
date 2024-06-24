---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_alias"
description: |-
  Get information on a AWS Key Management Service (KMS) Alias
---

# Data Source: aws_kms_alias

Use this data source to get the ARN of a KMS key alias.
By using this data source, you can reference key alias
without having to hard code the ARN as input.

## Example Usage

```terraform
data "aws_kms_alias" "s3" {
  name = "alias/aws/s3"
}
```

## Argument Reference

* `name` - (Required) Display name of the alias. The name must start with the word "alias" followed by a forward slash (alias/)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name(ARN) of the key alias.
* `id` - Amazon Resource Name(ARN) of the key alias.
* `target_key_id` - Key identifier pointed to by the alias.
* `target_key_arn` - ARN pointed to by the alias.
* `name` - Name of the alias
* `name_prefix` - Prefix of the alias
