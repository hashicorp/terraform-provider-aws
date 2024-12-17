---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate"
description: |-
  Provides a resource to issue a certificate using AWS Certificate Manager Private Certificate Authority (ACM PCA)
---

# Resource: aws_acmpca_certificate

Provides a resource to issue a certificate using AWS Certificate Manager Private Certificate Authority (ACM PCA).

Certificates created using `aws_acmpca_certificate` are not eligible for automatic renewal,
and must be replaced instead.
To issue a renewable certificate using an ACM PCA, create a [`aws_acm_certificate`](acm_certificate.html)
with the parameter `certificate_authority_arn`.

## Example Usage

### Basic

```terraform
resource "aws_acmpca_certificate" "example" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.example.arn
  certificate_signing_request = tls_cert_request.csr.cert_request_pem
  signing_algorithm           = "SHA256WITHRSA"
  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "example" {
  certificate_authority_configuration {
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
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `certificate_authority_arn` - (Required) ARN of the certificate authority.
* `certificate_signing_request` - (Required) Certificate Signing Request in PEM format.
* `signing_algorithm` - (Required) Algorithm to use to sign certificate requests. Valid values: `SHA256WITHRSA`, `SHA256WITHECDSA`, `SHA384WITHRSA`, `SHA384WITHECDSA`, `SHA512WITHRSA`, `SHA512WITHECDSA`.
* `validity` - (Required) Configures end of the validity period for the certificate. See [validity block](#validity-block) below.
* `template_arn` - (Optional) Template to use when issuing a certificate.
  See [ACM PCA Documentation](https://docs.aws.amazon.com/privateca/latest/userguide/UsingTemplates.html) for more information.
* `api_passthrough` - (Optional) Specifies X.509 certificate information to be included in the issued certificate. To use with API Passthrough templates

### validity block

* `type` - (Required) Determines how `value` is interpreted. Valid values: `DAYS`, `MONTHS`, `YEARS`, `ABSOLUTE`, `END_DATE`.
* `value` - (Required) If `type` is `DAYS`, `MONTHS`, or `YEARS`, the relative time until the certificate expires. If `type` is `ABSOLUTE`, the date in seconds since the Unix epoch. If `type` is `END_DATE`, the  date in RFC 3339 format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the certificate.
* `certificate` - PEM-encoded certificate value.
* `certificate_chain` - PEM-encoded certificate chain that includes any intermediate certificates and chains up to root CA.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ACM PCA Certificates using their ARN. For example:

```terraform
import {
  to = aws_acmpca_certificate.cert
  id = "arn:aws:acm-pca:eu-west-1:675225743824:certificate-authority/08319ede-83g9-1400-8f21-c7d12b2b6edb/certificate/a4e9c2aa4bcfab625g1b9136464cd3a"
}
```

Using `terraform import`, import ACM PCA Certificates using their ARN. For example:

```console
% terraform import aws_acmpca_certificate.cert arn:aws:acm-pca:eu-west-1:675225743824:certificate-authority/08319ede-83g9-1400-8f21-c7d12b2b6edb/certificate/a4e9c2aa4bcfab625g1b9136464cd3a
```
