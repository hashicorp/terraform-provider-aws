---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_alias"
description: |-
  Provides a display name for a customer master key.
---

# Resource: aws_kms_alias

Provides an alias for a KMS customer master key. AWS Console enforces 1-to-1 mapping between aliases & keys,
but API (hence Terraform too) allows you to create as many aliases as
the [account limits](http://docs.aws.amazon.com/kms/latest/developerguide/limits.html) allow you.

## Example Usage

```terraform
resource "aws_kms_key" "a" {}

resource "aws_kms_alias" "a" {
  name          = "alias/my-key-alias"
  target_key_id = aws_kms_key.a.key_id
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional) The display name of the alias. The name must start with the word "alias" followed by a forward slash (alias/)
* `name_prefix` - (Optional) Creates an unique alias beginning with the specified prefix.
The name must start with the word "alias" followed by a forward slash (alias/).  Conflicts with `name`.
* `target_key_id` - (Required) Identifier for the key for which the alias is for, can be either an ARN or key_id.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the key alias.
* `target_key_arn` - The Amazon Resource Name (ARN) of the target key identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import KMS aliases using the `name`. For example:

```terraform
import {
  to = aws_kms_alias.a
  id = "alias/my-key-alias"
}
```

Using `terraform import`, import KMS aliases using the `name`. For example:

```console
% terraform import aws_kms_alias.a alias/my-key-alias
```
