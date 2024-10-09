---
subcategory: "Payment Cryptography Control Plane"
layout: "aws"
page_title: "AWS: aws_paymentcryptography_key_alias"
description: |-
  Terraform resource for managing an AWS Payment Cryptography Control Plane Key Alias.
---
# Resource: aws_paymentcryptography_key_alias

Terraform resource for managing an AWS Payment Cryptography Control Plane Key Alias.

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

resource "aws_paymentcryptography_key_alias" "test" {
  alias_name = "alias/test-alias"
  key_arn    = aws_paymentcryptography_key.test.arn
}
```

## Argument Reference

The following arguments are required:

* `alias_name` - (Required) Name of the Key Alias.

The following arguments are optional:

* `key_arn` - (Optional) ARN of the key.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Payment Cryptography Control Plane Key Alias using the `alias/4681482429376900170`. For example:

```terraform
import {
  to = aws_paymentcryptography_key_alias.example
  id = "alias/4681482429376900170"
}
```

Using `terraform import`, import Payment Cryptography Control Plane Key Alias using the `alias/4681482429376900170`. For example:

```console
% terraform import aws_paymentcryptography_key_alias.example alias/4681482429376900170
```
