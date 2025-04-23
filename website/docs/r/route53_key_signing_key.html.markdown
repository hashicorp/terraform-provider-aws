---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_key_signing_key"
description: |-
    Manages an Route 53 Key Signing Key
---

# Resource: aws_route53_key_signing_key

Manages a Route 53 Key Signing Key. To manage Domain Name System Security Extensions (DNSSEC) for a Hosted Zone, see the [`aws_route53_hosted_zone_dnssec` resource](route53_hosted_zone_dnssec.html). For more information about managing DNSSEC in Route 53, see the [Route 53 Developer Guide](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec.html).

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
```

## Argument Reference

The following arguments are required:

* `hosted_zone_id` - (Required) Identifier of the Route 53 Hosted Zone.
* `key_management_service_arn` - (Required) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key. This must be unique for each key-signing key (KSK) in a single hosted zone. This key must be in the `us-east-1` Region and meet certain requirements, which are described in the [Route 53 Developer Guide](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec-cmk-requirements.html) and [Route 53 API Reference](https://docs.aws.amazon.com/Route53/latest/APIReference/API_CreateKeySigningKey.html).
* `name` - (Required) Name of the key-signing key (KSK). Must be unique for each key-signing key in the same hosted zone.

The following arguments are optional:

* `status` - (Optional) Status of the key-signing key (KSK). Valid values: `ACTIVE`, `INACTIVE`. Defaults to `ACTIVE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `digest_algorithm_mnemonic` - A string used to represent the delegation signer digest algorithm. This value must follow the guidelines provided by [RFC-8624 Section 3.3](https://tools.ietf.org/html/rfc8624#section-3.3).
* `digest_algorithm_type` - An integer used to represent the delegation signer digest algorithm. This value must follow the guidelines provided by [RFC-8624 Section 3.3](https://tools.ietf.org/html/rfc8624#section-3.3).
* `digest_value` - A cryptographic digest of a DNSKEY resource record (RR). DNSKEY records are used to publish the public key that resolvers can use to verify DNSSEC signatures that are used to secure certain kinds of information provided by the DNS system.
* `dnskey_record` - A string that represents a DNSKEY record.
* `ds_record` - A string that represents a delegation signer (DS) record.
* `flag` - An integer that specifies how the key is used. For key-signing key (KSK), this value is always 257.
* `id` - Route 53 Hosted Zone identifier and KMS Key identifier, separated by a comma (`,`).
* `key_tag` - An integer used to identify the DNSSEC record for the domain name. The process used to calculate the value is described in [RFC-4034 Appendix B](https://tools.ietf.org/rfc/rfc4034.txt).
* `public_key` - The public key, represented as a Base64 encoding, as required by [RFC-4034 Page 5](https://tools.ietf.org/rfc/rfc4034.txt).
* `signing_algorithm_mnemonic` - A string used to represent the signing algorithm. This value must follow the guidelines provided by [RFC-8624 Section 3.1](https://tools.ietf.org/html/rfc8624#section-3.1).
* `signing_algorithm_type` - An integer used to represent the signing algorithm. This value must follow the guidelines provided by [RFC-8624 Section 3.1](https://tools.ietf.org/html/rfc8624#section-3.1).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_route53_key_signing_key` resources using the Route 53 Hosted Zone identifier and KMS Key identifier, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_route53_key_signing_key.example
  id = "Z1D633PJN98FT9,example"
}
```

Using `terraform import`, import `aws_route53_key_signing_key` resources using the Route 53 Hosted Zone identifier and KMS Key identifier, separated by a comma (`,`). For example:

```console
% terraform import aws_route53_key_signing_key.example Z1D633PJN98FT9,example
```
