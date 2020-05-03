---
subcategory: "ACM PCA"
layout: "aws"
page_title: "AWS: aws_acmpca_private_certificate"
description: |-
  Get information on a AWS Certificate Manager Private Certificate Issuing
---

# Data Source: aws_acmpca_private_certificate

Get information on a AWS Certificate Manager Private Certificate Authority (ACM PCA Certificate Authority).

## Example Usage

```hcl
data "aws_acmpca_private_certificate" "example" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012/certificate/1234b4a0d73e2056789bdbe77d5b1a23"
  certificate_authority_arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) Amazon Resource Name (ARN) of the certificate issued by the private certificate authority.
* `certificate_authority_arn` - (Required) Amazon Resource Name (ARN) of the certificate authority.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the certificate authority.
* `certificate` - Base64-encoded certificate authority (CA) certificate. Only available after the certificate authority certificate has been imported.
* `certificate_chain` - Base64-encoded certificate chain that includes any intermediate certificates and chains up to root on-premises certificate that you used to sign your private CA certificate. The chain does not include your private CA certificate. Only available after the certificate authority certificate has been imported.
