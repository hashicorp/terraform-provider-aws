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

### CloudHSM

```terraform
resource "aws_kms_custom_key_store" "test" {
  cloud_hsm_cluster_id  = var.cloud_hsm_cluster_id
  custom_key_store_name = "kms-custom-key-store-test"
  key_store_password    = "noplaintextpasswords1"

  trust_anchor_certificate = file("anchor-certificate.crt")
}
```

### External Key Store (VPC)

```terraform
resource "aws_kms_custom_key_store" "example" {
  custom_key_store_name = "example-vpc-xks"
  custom_key_store_type = "EXTERNAL_KEY_STORE"

  xks_proxy_authentication_credential {
    access_key_id         = var.ephemeral_access_key_id
    raw_secret_access_key = var.ephemeral_secret_access_key
  }
  xks_proxy_connectivity              = "VPC_ENDPOINT_SERVICE"
  xks_proxy_uri_endpoint              = "https://myproxy-private.xks.example.com"
  xks_proxy_uri_path                  = "/kms/xks/v1"
  xks_proxy_vpc_endpoint_service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-example"
}
```

### External Key Store (Public)

```terraform
resource "aws_kms_custom_key_store" "example" {
  custom_key_store_name = "example-public-xks"
  custom_key_store_type = "EXTERNAL_KEY_STORE"

  xks_proxy_authentication_credential {
    access_key_id         = var.ephemeral_access_key_id
    raw_secret_access_key = var.ephemeral_secret_access_key
  }
  xks_proxy_connectivity = "PUBLIC_ENDPOINT"
  xks_proxy_uri_endpoint = "https://myproxy.xks.example.com"
  xks_proxy_uri_path     = "/kms/xks/v1"
}
```

## Argument Reference

The following arguments are required:

* `custom_key_store_name` - (Required) Unique name for Custom Key Store.

The following arguments are optional:

* `custom_key_store_type` - (Optional, ForceNew) Specifies the type of key store to create. Valid values are `AWS_CLOUDHSM` and `EXTERNAL_KEY_STORE`. If omitted, AWS will default the value to `AWS_CLOUDHSM`.

If `custom_key_store_type` is `AWS_CLOUDHSM`, the following optional arguments must be set:

* `cloud_hsm_cluster_id` - (Optional) Cluster ID of CloudHSM.
* `key_store_password` - (Optional) Specifies the `kmsuser` password for an AWS CloudHSM key store.
* `trust_anchor_certificate` - (Optional) Specifies the certificate for an AWS CloudHSM key store.

If `custom_key_store_type` is `EXTERNAL_KEY_STORE`, the following optional arguments must be set:

* `xks_proxy_authentication_credential` - (Optional) Specifies an authentication credential for the external key store proxy (XKS proxy). See [`xks_proxy_authentication_credential` attribute reference](#xks_proxy_authentication_credential-argument-reference) below.
* `xks_proxy_connectivity` - (Optional) Indicates how AWS KMS communicates with the external key store proxy.
* `xks_proxy_uri_endpoint` - (Optional) Specifies the endpoint that AWS KMS uses to send requests to the external key store proxy (XKS proxy).
* `xks_proxy_uri_path` - (Optional) Specifies the base path to the proxy APIs for this external key store. To find this value, see the documentation for your external key store proxy.
* `xks_proxy_vpc_endpoint_service_name` - (Optional) Specifies the name of the Amazon VPC endpoint service for interface endpoints that is used to communicate with your external key store proxy (XKS proxy). This argument is required when the value of `xks_proxy_connectivity` is `VPC_ENDPOINT_SERVICE`.

### `xks_proxy_authentication_credential` Argument Reference

* `access_key_id` - (Required) A unique identifier for the raw secret access key.
* `raw_secret_access_key` - (Required) A secret string of 43-64 characters.

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
