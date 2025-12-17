---
subcategory: "ACM (Certificate Manager)"
layout: "aws"
page_title: "AWS: aws_acm_certificate_export"
description: |-
  Exports an ACM certificate and private key
---

# Resource: aws_acm_certificate_export

Exports an existing ACM certificate and its private key. This resource calls the AWS ACM `ExportCertificate` API to retrieve the certificate, certificate chain, and encrypted private key.

This resource is useful for:

* Distributing ACM-issued certificates to on-premises or hybrid workloads
* Using ACM certificates outside of AWS services
* Managing certificate lifecycle entirely within Terraform without manual export steps

~> **NOTE:** Only certificates that are exportable can be used with this resource. For Amazon-issued certificates, the certificate must have been requested with exportable option enabled (available since June 2025). For private certificates issued by ACM Private CA, all certificates are exportable by default.

~> **NOTE:** The passphrase is required by the AWS API to encrypt the private key for secure export. The private key will be encrypted with the passphrase using AES-256-CBC encryption.

## Example Usage

### Export an Imported Certificate

```terraform
resource "aws_acm_certificate" "imported" {
  certificate_body = file("certificate.pem")
  private_key      = file("private_key.pem")

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_export" "example" {
  certificate_arn = aws_acm_certificate.imported.arn
  passphrase      = "my-secure-passphrase-123"
}

# Use the exported certificate in outputs
output "certificate" {
  value     = aws_acm_certificate_export.example.certificate
  sensitive = true
}

output "private_key" {
  value     = aws_acm_certificate_export.example.private_key
  sensitive = true
}
```

### Export a Private CA Certificate

```terraform
resource "aws_acm_certificate" "private_cert" {
  domain_name               = "example.com"
  certificate_authority_arn = aws_acmpca_certificate_authority.example.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_export" "private_cert" {
  certificate_arn = aws_acm_certificate.private_cert.arn
  passphrase      = var.certificate_passphrase
}
```

### Export and Store in AWS Secrets Manager

```terraform
resource "aws_acm_certificate" "example" {
  certificate_body = file("certificate.pem")
  private_key      = file("private_key.pem")
}

resource "aws_acm_certificate_export" "example" {
  certificate_arn = aws_acm_certificate.example.arn
  passphrase      = random_password.passphrase.result
}

resource "random_password" "passphrase" {
  length  = 32
  special = true
}

resource "aws_secretsmanager_secret" "cert" {
  name = "example-certificate"
}

resource "aws_secretsmanager_secret_version" "cert" {
  secret_id = aws_secretsmanager_secret.cert.id
  secret_string = jsonencode({
    certificate       = aws_acm_certificate_export.example.certificate
    certificate_chain = aws_acm_certificate_export.example.certificate_chain
    private_key       = aws_acm_certificate_export.example.private_key
    passphrase        = random_password.passphrase.result
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `certificate_arn` - (Required) ARN of the certificate to export. The certificate must be exportable.
* `passphrase` - (Required) Passphrase used to encrypt the private key. The private key is encrypted using AES-256-CBC. Must be at least 4 characters long.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the export (a hash of the certificate ARN and passphrase).
* `certificate` - PEM-encoded certificate.
* `certificate_chain` - PEM-encoded certificate chain. This is empty if the certificate was a self-signed certificate.
* `private_key` - Encrypted PEM-encoded private key. The private key is encrypted with the passphrase using AES-256-CBC encryption.

~> **NOTE:** All exported attributes (`certificate`, `certificate_chain`, and `private_key`) are marked as sensitive and will not be displayed in Terraform output unless explicitly configured to do so.

## Import

The ACM Certificate Export resource cannot be imported because it requires the passphrase to be provided at creation time, and the passphrase cannot be retrieved from AWS.

## Security Considerations

* The `passphrase` and all exported certificate materials are stored in the Terraform state file. Ensure your state file is properly secured and encrypted.
* Consider using [sensitive variables](https://www.terraform.io/language/values/variables#suppressing-values-in-cli-output) for the passphrase.
* Use [Terraform state encryption](https://www.terraform.io/language/state/encryption) to protect sensitive data at rest.
* Consider using [remote state backends](https://www.terraform.io/language/state/remote) with encryption enabled.
* The private key is encrypted by AWS using the provided passphrase before being returned by the API.

## Related Resources

* [`aws_acm_certificate`](acm_certificate.html) - Requests and manages ACM certificates
* [`aws_acm_certificate_validation`](acm_certificate_validation.html) - Waits for ACM certificate validation
* [`aws_secretsmanager_secret`](secretsmanager_secret.html) - Manages secrets in AWS Secrets Manager
