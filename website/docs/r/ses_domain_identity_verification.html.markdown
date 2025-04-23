---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_domain_identity_verification"
description: |-
  Waits for and checks successful verification of an SES domain identity.
---

# Resource: aws_ses_domain_identity_verification

Represents a successful verification of an SES domain identity.

Most commonly, this resource is used together with [`aws_route53_record`](route53_record.html) and
[`aws_ses_domain_identity`](ses_domain_identity.html) to request an SES domain identity,
deploy the required DNS verification records, and wait for verification to complete.

~> **WARNING:** This resource implements a part of the verification workflow. It does not represent a real-world entity in AWS, therefore changing or deleting this resource on its own has no immediate effect.

## Example Usage

```terraform
resource "aws_ses_domain_identity" "example" {
  domain = "example.com"
}

resource "aws_route53_record" "example_amazonses_verification_record" {
  zone_id = aws_route53_zone.example.zone_id
  name    = "_amazonses.${aws_ses_domain_identity.example.domain}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.example.verification_token]
}

resource "aws_ses_domain_identity_verification" "example_verification" {
  domain = aws_ses_domain_identity.example.domain

  depends_on = [aws_route53_record.example_amazonses_verification_record]
}
```

## Argument Reference

This resource supports the following arguments:

* `domain` - (Required) The domain name of the SES domain identity to verify.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The domain name of the domain identity.
* `arn` - The ARN of the domain identity.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `45m`)
