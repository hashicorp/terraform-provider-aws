---
subcategory: "Route 53 Domains"
layout: "aws"
page_title: "AWS: aws_route53domains_delegation_signer_record"
description: |-
  Provides a resource to manage a delegation signer record in the parent DNS zone for domains registered with Route53.
---

# Resource: aws_route53domains_delegation_signer_record

Provides a resource to manage a [delegation signer record](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec-enable-signing.html#dns-configuring-dnssec-enable-signing-step-1) in the parent DNS zone for domains registered with Route53.

## Example Usage

### Basic Usage

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
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
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
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = "example"
}

resource "aws_route53_hosted_zone_dnssec" "example" {
  depends_on = [
    aws_route53_key_signing_key.example
  ]
  hosted_zone_id = aws_route53_key_signing_key.example.hosted_zone_id
}

resource "aws_route53domains_delegation_signer_record" "example" {
  domain_name = "example.com"

  signing_attributes {
    algorithm  = aws_route53_key_signing_key.example.signing_algorithm_type
    flags      = aws_route53_key_signing_key.example.flag
    public_key = aws_route53_key_signing_key.example.public_key
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) The name of the domain that will have its parent DNS zone updated with the Delegation Signer record.
* `signing_attributes` - (Required) The information about a key, including the algorithm, public key-value, and flags.
    * `algorithm` - (Required) Algorithm which was used to generate the digest from the public key.
    * `flags` - (Required) Defines the type of key. It can be either a KSK (key-signing-key, value `257`) or ZSK (zone-signing-key, value `256`).
    * `public_key` - (Required) The base64-encoded public key part of the key pair that is passed to the registry.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `dnssec_key_id` - An ID assigned to the created DS record.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import delegation signer records using the domain name and DNSSEC key ID, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_route53domains_delegation_signer_record.example
  id = "example.com,40DE3534F5324DBDAC598ACEDB5B1E26A5368732D9C791D1347E4FBDDF6FC343"
}
```

Using `terraform import`, import delegation signer records using the domain name and DNSSEC key ID, separated by a comma (`,`). For example:

```console
% terraform import aws_route53domains_delegation_signer_record.example example.com,40DE3534F5324DBDAC598ACEDB5B1E26A5368732D9C791D1347E4FBDDF6FC343
```
