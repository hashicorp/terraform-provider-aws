---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_hosted_zone_dnssec"
description: |-
    Manages Route 53 Hosted Zone DNSSEC
---

# Resource: aws_route53_hosted_zone_dnssec

Manages Route 53 Hosted Zone Domain Name System Security Extensions (DNSSEC). For more information about managing DNSSEC in Route 53, see the [Route 53 Developer Guide](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec.html).

!> **WARNING:** If you disable DNSSEC signing for your hosted zone before the DNS changes have propagated, your domain could become unavailable on the internet. When you remove the DS records, you must wait until the longest TTL for the DS records that you remove has expired before you complete the step to disable DNSSEC signing. Please refer to the [Route 53 Developer Guide - Disable DNSSEC](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec-disable.html) for a detailed breakdown on the steps required to disable DNSSEC safely for a hosted zone.

## Example Usage

```terraform
provider "aws" {
  region = "us-east-1"
}

data "aws_caller_identity" "current" {}

resource "aws_kms_key" "example" {
  customer_master_key_spec = "ECC_NIST_P256"
  deletion_window_in_days  = 7
  key_usage                = "SIGN_VERIFY"
  policy = jsonencode({
    Statement = [
      {
        Action = [
          "kms:DescribeKey",
          "kms:GetPublicKey",
          "kms:Sign",
        ],
        Effect = "Allow"
        Principal = {
          Service = "dnssec-route53.amazonaws.com"
        }
        Sid      = "Allow Route 53 DNSSEC Service",
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = "${data.aws_caller_identity.current.account_id}"
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:route53:::hostedzone/*"
          }
        }
      },
      {
        Action = "kms:CreateGrant",
        Effect = "Allow"
        Principal = {
          Service = "dnssec-route53.amazonaws.com"
        }
        Sid      = "Allow Route 53 DNSSEC Service to CreateGrant",
        Resource = "*"
        Condition = {
          Bool = {
            "kms:GrantIsForAWSResource" = "true"
          }
        }
      },
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}

resource "aws_route53_zone" "example" {
  name = "example.com"
}

resource "aws_route53_key_signing_key" "example" {
  hosted_zone_id             = aws_route53_zone.example.id
  key_management_service_arn = aws_kms_key.example.arn
  name                       = "example"
}

resource "aws_route53_hosted_zone_dnssec" "example" {
  depends_on = [
    aws_route53_key_signing_key.example
  ]
  hosted_zone_id = aws_route53_key_signing_key.example.hosted_zone_id
}
```

## Argument Reference

The following arguments are required:

* `hosted_zone_id` - (Required) Identifier of the Route 53 Hosted Zone.

The following arguments are optional:

* `signing_status` - (Optional) Hosted Zone signing status. Valid values: `SIGNING`, `NOT_SIGNING`. Defaults to `SIGNING`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Route 53 Hosted Zone identifier.

## Import

`aws_route53_hosted_zone_dnssec` resources can be imported by using the Route 53 Hosted Zone identifier, e.g.,

```
$ terraform import aws_route53_hosted_zone_dnssec.example Z1D633PJN98FT9
```
