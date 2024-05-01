---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_signing_certificate"
description: |-
  Provides an IAM Signing Certificate
---

# Resource: aws_iam_signing_certificate

Provides an IAM Signing Certificate resource to upload Signing Certificates.

~> **Note:** All arguments including the certificate body will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

**Using certs on file:**

```terraform
resource "aws_iam_signing_certificate" "test_cert" {
  username         = "some_test_cert"
  certificate_body = file("self-ca-cert.pem")
}
```

**Example with cert in-line:**

```terraform
resource "aws_iam_signing_certificate" "test_cert_alt" {
  username = "some_test_cert"

  certificate_body = <<EOF
-----BEGIN CERTIFICATE-----
[......] # cert contents
-----END CERTIFICATE-----
EOF
}
```

## Argument Reference

This resource supports the following arguments:

* `certificate_body` – (Required) The contents of the signing certificate in PEM-encoded format.
* `status` – (Optional)  The status you want to assign to the certificate. `Active` means that the certificate can be used for programmatic calls to Amazon Web Services `Inactive` means that the certificate cannot be used.
* `user_name` – (Required) The name of the user the signing certificate is for.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `certificate_id` - The ID for the signing certificate.
* `id` - The `certificate_id:user_name`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Signing Certificates using the `id`. For example:

```terraform
import {
  to = aws_iam_signing_certificate.certificate
  id = "IDIDIDIDID:user-name"
}
```

Using `terraform import`, import IAM Signing Certificates using the `id`. For example:

```console
% terraform import aws_iam_signing_certificate.certificate IDIDIDIDID:user-name
```
