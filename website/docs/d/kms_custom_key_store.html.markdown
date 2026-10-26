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

This data source supports the following arguments:

* `custom_key_store_id` - (Optional) ID for the custom key store.
* `custom_key_store_name` - (Optional) User-specified friendly name for the custom key store.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cloud_hsm_cluster_id` - ID for the CloudHSM cluster that is associated with the custom key store.
* `connection_state` - Whether the custom key store is connected to its CloudHSM cluster.
* `creation_date` - Date and time when the custom key store was created.
* `id` - ID for the custom key store.
* `trust_anchor_certificate` - Trust anchor certificate of the associated CloudHSM cluster.
