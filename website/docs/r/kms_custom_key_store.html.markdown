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

* `cloud_hsm_cluster_id` - (Optional) Cluster ID of CloudHSM. Applies when `custom_key_store_type` is `AWS_CLOUDHSM`.
* `custom_key_store_type` - (Optional, ForceNew) Type of key store to create. Valid values are `AWS_CLOUDHSM` and `EXTERNAL_KEY_STORE`. If omitted, AWS will default the value to `AWS_CLOUDHSM`.
* `key_store_password` - (Optional) Password for the `kmsuser` of an AWS CloudHSM key store. Applies when `custom_key_store_type` is `AWS_CLOUDHSM`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `trust_anchor_certificate` - (Optional) Certificate for an AWS CloudHSM key store. Applies when `custom_key_store_type` is `AWS_CLOUDHSM`.
* `xks_proxy_authentication_credential` - (Optional) Authentication credential for the external key store proxy (XKS proxy). Applies when `custom_key_store_type` is `EXTERNAL_KEY_STORE`. See [`xks_proxy_authentication_credential` block](#xks_proxy_authentication_credential-block) below.
* `xks_proxy_connectivity` - (Optional) How AWS KMS communicates with the external key store proxy. Applies when `custom_key_store_type` is `EXTERNAL_KEY_STORE`.
* `xks_proxy_uri_endpoint` - (Optional) Endpoint that AWS KMS uses to send requests to the external key store proxy (XKS proxy). Applies when `custom_key_store_type` is `EXTERNAL_KEY_STORE`.
* `xks_proxy_uri_path` - (Optional) Base path to the proxy APIs for this external key store. To find this value, see the documentation for your external key store proxy. Applies when `custom_key_store_type` is `EXTERNAL_KEY_STORE`.
* `xks_proxy_vpc_endpoint_service_name` - (Optional) Name of the Amazon VPC endpoint service for interface endpoints that is used to communicate with your external key store proxy (XKS proxy). Applies when `custom_key_store_type` is `EXTERNAL_KEY_STORE` and required when `xks_proxy_connectivity` is `VPC_ENDPOINT_SERVICE`.

### `xks_proxy_authentication_credential` Block

* `access_key_id` - (Required) Unique identifier for the raw secret access key.
* `raw_secret_access_key` - (Required) Secret string of 43-64 characters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Custom Key Store ID.

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
