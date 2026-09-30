---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_key"
description: |-
  Get information on a AWS Key Management Service (KMS) Key
---

# Data Source: aws_kms_key

Use this data source to get detailed information about
the specified KMS Key with flexible key id input.
This can be useful to reference key alias
without having to hard code the ARN as input.

## Example Usage

```terraform
data "aws_kms_key" "by_alias" {
  key_id = "alias/my-key"
}

data "aws_kms_key" "by_id" {
  key_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}

data "aws_kms_key" "by_alias_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:alias/my-key"
}

data "aws_kms_key" "by_key_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

## Argument Reference

This data source supports the following arguments:

* `grant_tokens` - (Optional) List of grant tokens.
* `key_id` - (Required) Key identifier. Can be a key ID (e.g. `1234abcd-12ab-34cd-56ef-1234567890ab`), key ARN (e.g. `arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab`), alias name (e.g. `alias/my-key`), or alias ARN (e.g. `arn:aws:kms:us-east-1:111122223333:alias/my-key`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the key.
* `aws_account_id` - Twelve-digit account ID of the AWS account that owns the key.
* `cloud_hsm_cluster_id` - Cluster ID of the AWS CloudHSM cluster that contains the key material for the KMS key.
* `creation_date` - Date and time when the key was created.
* `custom_key_store_id` - Unique identifier for the custom key store that contains the KMS key.
* `customer_master_key_spec` - See `key_spec`.
* `deletion_date` - Date and time after which AWS KMS deletes the key. This value is present only when `key_state` is `PendingDeletion`, otherwise this value is 0.
* `description` - Description of the key.
* `enabled` - Whether the key is enabled. When `key_state` is `Enabled` this value is true, otherwise it is false.
* `expiration_model` - Whether the key's key material expires. This value is present only when `origin` is `EXTERNAL`, otherwise this value is empty.
* `id` - Globally unique identifier for the key.
* `key_manager` - Manager of the key.
* `key_spec` - Type of key material in the KMS key.
* `key_state` - State of the key.
* `key_usage` - Intended use of the key.
* `multi_region` - Whether the KMS key is a multi-Region (`true`) or regional (`false`) key.
* `multi_region_configuration` - Primary and replica keys in same multi-Region key. Present only when the value of `multi_region` is `true`. See [`multi_region_configuration` Block](#multi_region_configuration-block) below.
* `origin` - Source of the key material. When this value is `AWS_KMS`, AWS KMS created the key material. When this value is `EXTERNAL`, the key material was imported from your existing key management infrastructure or the CMK lacks key material.
* `pending_deletion_window_in_days` - Waiting period before the primary key in a multi-Region key is deleted.
* `valid_to` - Time at which the imported key material expires. This value is present only when `origin` is `EXTERNAL` and whose `expiration_model` is `KEY_MATERIAL_EXPIRES`, otherwise this value is 0.
* `xks_key_configuration` - Information about the external key that is associated with a KMS key in an external key store. See [`xks_key_configuration` Block](#xks_key_configuration-block) below.

### `multi_region_configuration` Block

The `multi_region_configuration` block contains:

* `multi_region_key_type` - Whether the KMS key is a `PRIMARY` or `REPLICA` key.
* `primary_key` - Key ARN and Region of the primary key. This is the current KMS key if it is the primary key. See [`multi_region_configuration.primary_key` Block](#multi_region_configurationprimary_key-block) below.
* `replica_keys` - Key ARNs and Regions of all replica keys. Includes the current KMS key if it is a replica key. See [`multi_region_configuration.replica_keys` Block](#multi_region_configurationreplica_keys-block) below.

### `multi_region_configuration.primary_key` Block

The `multi_region_configuration.primary_key` block contains:

* `arn` - Key ARN of a primary or replica key of a multi-Region key.
* `region` - AWS Region of a primary or replica key in a multi-Region key.

### `multi_region_configuration.replica_keys` Block

The `multi_region_configuration.replica_keys` block contains:

* `arn` - Key ARN of a primary or replica key of a multi-Region key.
* `region` - AWS Region of a primary or replica key in a multi-Region key.

### `xks_key_configuration` Block

The `xks_key_configuration` block contains:

* `id` - ID of the external key in the external key manager.
