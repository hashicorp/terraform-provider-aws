---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_custom_key_store"
description: |-
  Terraform resource for managing an AWS KMS (Key Management) Custom Key Store.
---

# Resource: aws_kms_custom_key_store

Terraform resource for managing an AWS KMS (Key Management) Custom Key Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_custom_key_store" "test" {
  cloud_hsm_cluster_id  = var.cloud_hsm_cluster_id
  custom_key_store_name = "kms-custom-key-store-test"
  key_store_password    = "noplaintextpasswords1"

  trust_anchor_certificate = file("anchor-certificate.crt")
}
```

## Argument Reference

The following arguments are required:

* `cloud_hsm_cluster_id` - (Required) Cluster ID of CloudHSM.
* `custom_key_store_name` - (Required) Unique name for Custom Key Store.
* `key_store_password` - (Required) Password for `kmsuser` on CloudHSM.
* `trust_anchor_certificate` - (Required) Customer certificate used for signing on CloudHSM.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Custom Key Store ID

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import KMS (Key Management) Custom Key Store using the `id`. For example:

```terraform
import {
  to = aws_kms_custom_key_store.example
  id = "cks-5ebd4ef395a96288e"
}
```

Using `terraform import`, import KMS (Key Management) Custom Key Store using the `id`. For example:

```console
% terraform import aws_kms_custom_key_store.example cks-5ebd4ef395a96288e
```
