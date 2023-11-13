---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_policy"
description: |-
  Attaches a resource based policy to an AWS Certificate Manager Private Certificate Authority (ACM PCA)
---

# Resource: aws_acmpca_policy

Attaches a resource based policy to a private CA.

## Example Usage

### Basic

```terraform
data "aws_iam_policy_document" "example" {
  statement {
    sid    = "1"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }

    actions = [
      "acm-pca:DescribeCertificateAuthority",
      "acm-pca:GetCertificate",
      "acm-pca:GetCertificateAuthorityCertificate",
      "acm-pca:ListPermissions",
      "acm-pca:ListTags",
    ]

    resources = [aws_acmpca_certificate_authority.example.arn]
  }

  statement {
    sid    = "2"
    effect = Allow

    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }

    actions   = ["acm-pca:IssueCertificate"]
    resources = [aws_acmpca_certificate_authority.example.arn]

    condition {
      test     = "StringEquals"
      variable = "acm-pca:TemplateArn"
      values   = ["arn:aws:acm-pca:::template/EndEntityCertificate/V1"]
    }
  }
}

resource "aws_acmpca_policy" "example" {
  resource_arn = aws_acmpca_certificate_authority.example.arn
  policy       = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) ARN of the private CA to associate with the policy.
* `policy` - (Required) JSON-formatted IAM policy to attach to the specified private CA resource.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_acmpca_policy` using the `resource_arn` value. For example:

```terraform
import {
  to = aws_acmpca_policy.example
  id = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import `aws_acmpca_policy` using the `resource_arn` value. For example:

```console
% terraform import aws_acmpca_policy.example arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012
```
