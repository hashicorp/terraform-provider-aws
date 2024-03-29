---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_replica_external_key"
description: |-
  Manages a KMS multi-Region replica key that uses external key material.
---

# Resource: aws_kms_replica_external_key

Manages a KMS multi-Region replica key that uses external key material.
See the [AWS KMS Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/multi-region-keys-import.html) for more information on importing key material into multi-Region keys.

## Example Usage

```terraform
provider "aws" {
  alias  = "primary"
  region = "us-east-1"
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_kms_external_key" "primary" {
  provider = aws.primary

  description             = "Multi-Region primary key"
  deletion_window_in_days = 30
  multi_region            = true
  enabled                 = true

  key_material_base64 = "..."
}

resource "aws_kms_replica_external_key" "replica" {
  description             = "Multi-Region replica key"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_external.primary.arn

  key_material_base64 = "..." # Must be the same key material as the primary's.
}
```

## Argument Reference

This resource supports the following arguments:

* `bypass_policy_lockout_safety_check` - (Optional) A flag to indicate whether to bypass the key policy lockout safety check.
Setting this value to true increases the risk that the KMS key becomes unmanageable. Do not set this value to true indiscriminately.
For more information, refer to the scenario in the [Default Key Policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html#key-policy-default-allow-root-enable-iam) section in the _AWS Key Management Service Developer Guide_.
The default value is `false`.
* `deletion_window_in_days` - (Optional) The waiting period, specified in number of days. After the waiting period ends, AWS KMS deletes the KMS key.
If you specify a value, it must be between `7` and `30`, inclusive. If you do not specify a value, it defaults to `30`.
* `description` - (Optional) A description of the KMS key.
* `enabled` - (Optional) Specifies whether the replica key is enabled. Disabled KMS keys cannot be used in cryptographic operations. Keys pending import can only be `false`. Imported keys default to `true` unless expired.
* `key_material_base64` - (Optional) Base64 encoded 256-bit symmetric encryption key material to import. The KMS key is permanently associated with this key material. The same key material can be [reimported](https://docs.aws.amazon.com/kms/latest/developerguide/importing-keys.html#reimport-key-material), but you cannot import different key material.
* `policy` - (Optional) The key policy to attach to the KMS key. If you do not specify a key policy, AWS KMS attaches the [default key policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html#key-policy-default) to the KMS key.
For more information about building policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `primary_key_arn` - (Required) The ARN of the multi-Region primary key to replicate. The primary key must be in a different AWS Region of the same AWS Partition. You can create only one replica of a given primary key in each AWS Region.
* `tags` - (Optional) A map of tags to assign to the replica key. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `valid_to` - (Optional) Time at which the imported key material expires. When the key material expires, AWS KMS deletes the key material and the key becomes unusable. If not specified, key material does not expire. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the replica key. The key ARNs of related multi-Region keys differ only in the Region value.
* `expiration_model` - Whether the key material expires. Empty when pending key material import, otherwise `KEY_MATERIAL_EXPIRES` or `KEY_MATERIAL_DOES_NOT_EXPIRE`.
* `key_id` - The key ID of the replica key. Related multi-Region keys have the same key ID.
* `key_state` - The state of the replica key.
* `key_usage` - The [cryptographic operations](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#cryptographic-operations) for which you can use the KMS key. This is a shared property of multi-Region keys.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import KMS multi-Region replica keys using the `id`. For example:

```terraform
import {
  to = aws_kms_replica_external_key.example
  id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

Using `terraform import`, import KMS multi-Region replica keys using the `id`. For example:

```console
% terraform import aws_kms_replica_external_key.example 1234abcd-12ab-34cd-56ef-1234567890ab
```
