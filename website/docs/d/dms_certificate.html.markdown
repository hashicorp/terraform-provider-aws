---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_certificate"
description: |-
  Terraform data source for managing an AWS DMS (Database Migration) Certificate.
---

# Data Source: aws_dms_certificate

Terraform data source for managing an AWS DMS (Database Migration) Certificate.

## Example Usage

### Basic Usage

```terraform
data "aws_dms_certificate" "example" {
  certificate_id = aws_dms_certificate.test.certificate_id
}
```

## Argument Reference

The following arguments are required:

* `certificate_id` - (Required) A customer-assigned name for the certificate. Identifiers must begin with a letter and must contain only ASCII letters, digits, and hyphens. They can't end with a hyphen or contain two consecutive hyphens.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `certificate_creation_date` - The date that the certificate was created.
* `certificate_pem` - The contents of a .pem file, which contains an X.509 certificate.
* `certificate_owner` - The owner of the certificate.
* `certificate_arn` - The Amazon Resource Name (ARN) for the certificate.
* `certificate_wallet` - The owner of the certificate.
* `key_length` - The key length of the cryptographic algorithm being used.
* `signing_algorithm` - The algorithm for the certificate.
* `valid_from_date` - The beginning date that the certificate is valid.
* `valid_to_date` - The final date that the certificate is valid.
