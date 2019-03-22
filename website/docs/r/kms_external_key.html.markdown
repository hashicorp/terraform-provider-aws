---
layout: "aws"
page_title: "AWS: aws_kms_key"
sidebar_current: "docs-aws-resource-kms-key"
description: |-
  Provides a KMS customer master key that uses external key data
---

# aws\_kms\_key

Provides a KMS customer master key that uses external key data

~> **NOTE:** By default, AWS KMS creates key material for you when you create a customer master key (CMK). To instead import your own key material, you need to create a CMK with no key material, which is what this resource does. Currently importing the key material is not implemented in Terraform, and should be done as documented in the [AWS importing key material guide](https://docs.aws.amazon.com/kms/latest/developerguide/importing-keys.html)


## Example Usage

```hcl
resource "aws_kms_external_key" "a" {
  description = "KMS EXTERNAL for AMI encryption"
  deletion_window_in_days = 10
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the key as viewed in AWS console.
* `key_usage` - (Optional) Specifies the intended use of the key.
  Defaults to ENCRYPT/DECRYPT, and only symmetric encryption and decryption are supported.
* `policy` - (Optional) A valid policy JSON document.
* `deletion_window_in_days` - (Optional) Duration in days after which the key is deleted
  after destruction of the resource, must be between 7 and 30 days. Defaults to 30 days.
* `is_enabled` - (Optional) Specifies whether the key is enabled. this option is always false while `key_state` is `PendingImport`. Defaults to true.
* `tags` - (Optional) A mapping of tags to assign to the object.

## Attributes Reference

The following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the key.
* `key_id` - The globally unique identifier for the key.
* `key_state` - The state of the CMK.

## Import

KMS Keys can be imported using the `id`, e.g.

```
$ terraform import aws_kms_external_key.a arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
```
