---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_certificate"
description: |-
    Creates and manages an AWS IoT certificate.
---

# Resource: aws_iot_certificate

Creates and manages an AWS IoT certificate.

## Example Usage

### With CSR

```terraform
resource "aws_iot_certificate" "cert" {
  csr    = file("/my/csr.pem")
  active = true
}
```

### Without CSR

```terraform
resource "aws_iot_certificate" "cert" {
  active = true
}
```

### From existing certificate without a CA

```terraform
resource "aws_iot_certificate" "cert" {
  certificate_pem = file("/my/cert.pem")
  active          = true
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `active` - (Required)  Boolean flag to indicate if the certificate should be active
* `csr` - (Optional) The certificate signing request. Review
  [CreateCertificateFromCsr](https://docs.aws.amazon.com/iot/latest/apireference/API_CreateCertificateFromCsr.html)
  for more information on generating a certificate from a certificate signing request (CSR).
  If none is specified both the certificate and keys will be generated, review [CreateKeysAndCertificate](https://docs.aws.amazon.com/iot/latest/apireference/API_CreateKeysAndCertificate.html)
  for more information on generating keys and a certificate.
* `certificate_pem` - (Optional) The certificate to be registered. If `ca_pem` is unspecified, review
  [RegisterCertificateWithoutCA](https://docs.aws.amazon.com/iot/latest/apireference/API_RegisterCertificateWithoutCA.html).
  If `ca_pem` is specified, review
  [RegisterCertificate](https://docs.aws.amazon.com/iot/latest/apireference/API_RegisterCertificate.html)
  for more information on registering a certificate.
* `ca_pem` - (Optional) The CA certificate for the certificate to be registered. If this is set, the CA needs to be registered with AWS IoT beforehand.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The internal ID assigned to this certificate.
* `arn` - The ARN of the created certificate.
* `ca_certificate_id` - The certificate ID of the CA certificate used to sign the certificate.
* `certificate_pem` - The certificate data, in PEM format.
* `public_key` - When neither CSR nor certificate is provided, the public key.
* `private_key` - When neither CSR nor certificate is provided, the private key.
