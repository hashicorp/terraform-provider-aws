---
subcategory: "Payment Cryptography Control Plane"
layout: "aws"
page_title: "AWS: aws_paymentcryptography_key"
description: |-
  Terraform resource for managing an AWS Payment Cryptography Control Plane Key.
---
# Resource: aws_paymentcryptography_key

Terraform resource for managing an AWS Payment Cryptography Control Plane Key.

## Example Usage

### Basic Usage

```terraform
resource "aws_paymentcryptography_key" "test" {
  exportable = true
  key_attributes {
    key_algorithm = "TDES_3KEY"
    key_class     = "SYMMETRIC_KEY"
    key_usage     = "TR31_P0_PIN_ENCRYPTION_KEY"
    key_modes_of_use {
      decrypt = true
      encrypt = true
      wrap    = true
      unwrap  = true
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `exportable` - (Required) Whether the key is exportable from the service.
* `key_attributes` - (Required) Role of the key, the algorithm it supports, and the cryptographic operations allowed with the key.

The following arguments are optional:

* `enabled` - (Optional) Whether to enable the key.
* `key_check_value_algorithm` - (Optional) Algorithm that AWS Payment Cryptography uses to calculate the key check value (KCV).
* `tags` - (Optional) Map of tags assigned to the WorkSpaces Connection Alias. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### key_attributes

The following arguments are required:

* `key_algorithm` - (Required) Key algorithm to be use during creation of an AWS Payment Cryptography key.
* `key_class` - (Required) Type of AWS Payment Cryptography key to create.
* `key_modes_of_use`- (Required) List of cryptographic operations that you can perform using the key.
* `key_usage` - (Required) Cryptographic usage of an AWS Payment Cryptography key as defined in section A.5.2 of the TR-31 spec.

### key_modes_of_use

The following arguments are optional:

* `decrypt` - (Optional) Whether an AWS Payment Cryptography key can be used to decrypt data.
* `derive_key` - (Optional) Whether an AWS Payment Cryptography key can be used to derive new keys.
* `encrypt` - (Optional) Whether an AWS Payment Cryptography key can be used to encrypt data.
* `generate` - (Optional) Whether an AWS Payment Cryptography key can be used to generate and verify other card and PIN verification keys.
* `no_restrictions` - (Optional) Whether an AWS Payment Cryptography key has no special restrictions other than the restrictions implied by KeyUsage.
* `sign` - (Optional) Whether an AWS Payment Cryptography key can be used for signing.
* `unwrap` - (Optional) Whether an AWS Payment Cryptography key can be used to unwrap other keys.
* `verify` - (Optional) Whether an AWS Payment Cryptography key can be used to verify signatures.
* `wrap` - (Optional) Whether an AWS Payment Cryptography key can be used to wrap other keys.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the key.
* `key_check_value` - Key check value (KCV) is used to check if all parties holding a given key have the same key or to detect that a key has changed.
* `key_origin` - Source of the key material.
* `key_state` - State of key that is being created or deleted.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Payment Cryptography Control Plane Key using the `arn:aws:payment-cryptography:us-east-1:123456789012:key/qtbojf64yshyvyzf`. For example:

```terraform
import {
  to = aws_paymentcryptography_key.example
  id = "arn:aws:payment-cryptography:us-east-1:123456789012:key/qtbojf64yshyvyzf"
}
```

Using `terraform import`, import Payment Cryptography Control Plane Key using the `arn:aws:payment-cryptography:us-east-1:123456789012:key/qtbojf64yshyvyzf`. For example:

```console
% terraform import aws_paymentcryptography_key.example arn:aws:payment-cryptography:us-east-1:123456789012:key/qtbojf64yshyvyzf
```
