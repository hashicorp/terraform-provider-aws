---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_custom_key_store"
description: |-
  Get information on a AWS Key Management Service (KMS) Custom Key Store
---

# Data Source: aws_kms_custom_key_store

Use this data source to get the metadata KMS custom key store.
By using this data source, you can reference KMS custom key store
without having to hard code the ID as input.

## Example Usage

```terraform
data "aws_kms_custom_key_store" "keystore" {
  custom_key_store_name = "my_cloudhsm"
}
```

## Argument Reference

* `custom_key_store_id` - (Optional) The ID for the custom key store.
* `custom_key_store_name` - (Optional) The user-specified friendly name for the custom key store.

## Attributes Reference

* `id` - The ID for the custom key store.
* `cloudhsm_cluster_id` - ID for the CloudHSM cluster that is associated with the custom key store.
* `connection_state` - Indicates whether the custom key store is connected to its CloudHSM cluster.
* `creation_date` - The date and time when the custom key store was created.
* `trust_anchor_certificate` - The trust anchor certificate of the associated CloudHSM cluster.
