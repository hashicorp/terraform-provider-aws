---
subcategory: "ACM PCA"
layout: "aws"
page_title: "AWS: aws_acmpca_private_certificate"
description: |-
  Provides a resource to manage AWS Certificate Manager Private Certificate Issuing
---

# Resource: aws_acmpca_private_certificate

Provides a resource to manage AWS Certificate Manager Private Certificate Issuing (ACM PCA Certificate Issuing). It is to be used when signing a certificate that was generated outside AWS and have `aws_acmpca_certificate_authority` as the CA.

## Example Usage

### Basic

```hcl
resource "aws_acmpca_certificate_authority" "example" {
  private_certificate_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  permanent_deletion_time_in_days = 7
}

resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_cert_request" "csr" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example"
  }
}

resource "aws_acmpca_private_certificate" "example" {
	certificate_authority_arn = aws_acmpca_certificate_authority.example.arn
	certificate_signing_request = tls_cert_request.csr.cert_request_pem
	signing_algorithm = "SHA256WITHRSA"
	validity_length = 1
	validity_unit = "YEARS"
}
```

## Argument Reference

The following arguments are supported:

* `certificate_authority_arn` - (Required) Amazon Resource Name (ARN) of the certificate authority.
* `certificate_signing_request` - (Required) Certificate Signing Request in PEM format.
* `signing_algorithm` - (Required) Algorithm to use to sign certificate requests. Valid values: `SHA256WITHRSA`, `SHA256WITHECDSA`, `SHA384WITHRSA`, `SHA384WITHECDSA`, `SHA512WITHRSA`, `SHA512WITHECDSA`
* `validity_length` - (Required) Used with `validity_unit` as the number to apply with the unit
* `validity_unit` - (Required) The unit of time for certificate validity. Cannot be set to a value higher than the validity of the Certficate Authority. Valid values: `DAYS`, `MONTHS`, `YEARS`, `ABSOLUTE`, `END_DATE`.
* `template_arn` - (Optional) The template to use when issuing a certificate. See [ACM PCA Documentation](https://docs.aws.amazon.com/acm-pca/latest/userguide/UsingTemplates.html) for more information.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the certificate.
* `arn` - Amazon Resource Name (ARN) of the certificate.
* `certificate` - Certificate PEM.
* `certificate_chain` - Certificate chain that includes any intermediate certificates and chains up to root CA that you used to sign your private certificate.

## Timeouts

`aws_acmpca_private_certificate` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `5m`) How long to wait for a certificate authority to issue a certificate.

## Import

`aws_acmpca_private_certificate` can not be imported at this time.
