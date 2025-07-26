---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate"
description: |-
  Get information on a Certificate issued by a AWS Certificate Manager Private Certificate Authority
---

# Data Source: aws_acmpca_certificate

Get information on a Certificate issued by a AWS Certificate Manager Private Certificate Authority.

## Example Usage

```terraform
data "aws_acmpca_certificate" "example" {
  arn                       = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012/certificate/1234b4a0d73e2056789bdbe77d5b1a23"
  certificate_authority_arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the certificate issued by the private certificate authority.
* `certificate_authority_arn` - (Required) ARN of the certificate authority.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `certificate` - PEM-encoded certificate value.
* `certificate_chain` - PEM-encoded certificate chain that includes any intermediate certificates and chains up to root CA.
