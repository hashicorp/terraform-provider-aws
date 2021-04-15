---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_public_key"
description: |-
  Get public key on a AWS Key Management Service (KMS) Key
---

# aws_kms_public_key

Use this data source to get the public key about
the specified KMS Key with flexible key id input.
This can be useful to reference key alias
without having to hard code the ARN as input.

## Example Usage

```terraform
data "aws_kms_public_key" "by_alias" {
  key_id = "alias/my-key"
}

data "aws_kms_public_key" "by_id" {
  key_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}

data "aws_kms_public_key" "by_alias_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:alias/my-key"
}

data "aws_kms_public_key" "by_key_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

## Argument Reference

* `key_id` - (Required) Key identifier which can be one of the following format:
    * Key ID. E.g: `1234abcd-12ab-34cd-56ef-1234567890ab`
    * Key ARN. E.g.: `arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab`
    * Alias name. E.g.: `alias/my-key`
    * Alias ARN: E.g.: `arn:aws:kms:us-east-1:111122223333:alias/my-key`
* `grant_tokens` - (Optional) List of grant tokens

## Attributes Reference

* `id`: The globally unique identifier for the key
* `customer_master_key_spec`: The type of the public key that was downloaded
* `encryption_algorithms`: The encryption algorithms that AWS KMS supports for this key
* `key_usage`: The permitted use of the public key. Valid values are `ENCRYPT_DECRYPT` or `SIGN_VERIFY`
* `signing_algorithms`: The signing algorithms that AWS KMS supports for this key
* `public_key`: The exported public key in a DER-encoded X.509 string format