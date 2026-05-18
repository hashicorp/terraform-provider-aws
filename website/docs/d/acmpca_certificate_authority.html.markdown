---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate_authority"
description: |-
  Get information on a AWS Certificate Manager Private Certificate Authority
---

# Data Source: aws_acmpca_certificate_authority

Get information on a AWS Certificate Manager Private Certificate Authority (ACM PCA Certificate Authority).

## Example Usage

```terraform
data "aws_acmpca_certificate_authority" "example" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the certificate authority.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN of the certificate authority.
* `certificate` - Base64-encoded certificate authority (CA) certificate. Only available after the certificate authority certificate has been imported.
* `certificate_chain` - Base64-encoded certificate chain that includes any intermediate certificates and chains up to root on-premises certificate that you used to sign your private CA certificate. The chain does not include your private CA certificate. Only available after the certificate authority certificate has been imported.
* `certificate_signing_request` - The base64 PEM-encoded certificate signing request (CSR) for your private CA certificate.
* `usage_mode` - Specifies whether the CA issues general-purpose certificates that typically require a revocation mechanism, or short-lived certificates that may optionally omit revocation because they expire quickly.
* `key_storage_security_standard` - Level of security of the key storage endpoint of the certificate authority.
* `not_after` - Date and time after which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `not_before` - Date and time before which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `revocation_configuration` - Nested attribute containing revocation configuration. See [`revocation_configuration`](#revocation_configuration) below.

### `revocation_configuration`

* `crl_configuration` - Nested attribute containing configuration of the certificate revocation list (CRL). See [`crl_configuration`](#crl_configuration) below.
* `ocsp_configuration` - Nested attribute containing configuration of the Online Certificate Status Protocol (OCSP). See [`ocsp_configuration`](#ocsp_configuration) below.

### `crl_configuration`

* `custom_cname` - Name inserted into the certificate CRL Distribution Points extension that enables the use of an alias for the CRL distribution point.
* `custom_path` - Custom path for the CRL in S3.
* `enabled` - Boolean value that specifies whether certificate revocation lists (CRLs) are enabled.
* `expiration_in_days` - Number of days until a certificate expires.
* `s3_bucket_name` - Name of the S3 bucket that contains the CRL.
* `s3_object_acl` - Whether the CRL is publicly readable or privately held in the CRL Amazon S3 bucket.

### `ocsp_configuration`

* `enabled` - Boolean value that specifies whether a custom OCSP responder is enabled.
* `ocsp_custom_cname` - A CNAME specifying a customized OCSP domain.
* `serial` - Serial number of the certificate authority. Only available after the certificate authority certificate has been imported.
* `status` - Status of the certificate authority.
* `tags` - Key-value map of user-defined tags that are attached to the certificate authority.
* `type` - Type of the certificate authority.
