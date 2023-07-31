---
subcategory: "Roles Anywhere"
layout: "aws"
page_title: "AWS: aws_rolesanywhere_trust_anchor"
description: |-
  Provides a Roles Anywhere Trust Anchor resource
---

# Resource: aws_rolesanywhere_trust_anchor

Terraform resource for managing a Roles Anywhere Trust Anchor.

## Example Usage

```terraform
resource "aws_acmpca_certificate_authority" "example" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "example.com"
    }
  }
}

data "aws_partition" "current" {}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.example.arn
  certificate_signing_request = aws_acmpca_certificate_authority.example.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority_certificate" "example" {
  certificate_authority_arn = aws_acmpca_certificate_authority.example.arn
  certificate               = aws_acmpca_certificate.example.certificate
  certificate_chain         = aws_acmpca_certificate.example.certificate_chain
}

resource "aws_rolesanywhere_trust_anchor" "test" {
  name = "example"
  source {
    source_data {
      acm_pca_arn = aws_acmpca_certificate_authority.example.arn
    }
    source_type = "AWS_ACM_PCA"
  }
  # Wait for the ACMPCA to be ready to receive requests before setting up the trust anchor
  depends_on = [aws_acmpca_certificate_authority_certificate.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `enabled` - (Optional) Whether or not the Trust Anchor should be enabled.
* `name` - (Required) The name of the Trust Anchor.
* `source` - (Required) The source of trust, documented below
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested Blocks

#### `source`

* `source_data` - (Required) The data denoting the source of trust, documented below
* `source_type` - (Required) The type of the source of trust. Must be either `AWS_ACM_PCA` or `CERTIFICATE_BUNDLE`.

#### `source_data`

* `acm_pca_arn` - (Optional, required when `source_type` is `AWS_ACM_PCA`) The ARN of an ACM Private Certificate Authority.
* `x509_certificate_data` - (Optional, required when `source_type` is `CERTIFICATE_BUNDLE`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Trust Anchor
* `id` - The Trust Anchor ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_rolesanywhere_trust_anchor` using its `id`. For example:

```terraform
import {
  to = aws_rolesanywhere_trust_anchor.example
  id = "92b2fbbb-984d-41a3-a765-e3cbdb69ebb1"
}
```

Using `terraform import`, import `aws_rolesanywhere_trust_anchor` using its `id`. For example:

```console
% terraform import aws_rolesanywhere_trust_anchor.example 92b2fbbb-984d-41a3-a765-e3cbdb69ebb1
```
