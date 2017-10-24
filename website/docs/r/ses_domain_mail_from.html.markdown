---
layout: "aws"
page_title: "AWS: ses_domain_mail_from"
sidebar_current: "docs-aws-resource-ses-domain-mail-from"
description: |-
  Provides an SES domain MAIL FROM resource
---

# aws\_ses\_domain\_dkim

Provides an SES domain MAIL FROM resource.

Domain ownership needs to be confirmed first using [ses_domain_identity Resource](/docs/providers/aws/r/ses_domain_identity.html)

## Argument Reference

The following arguments are supported:

* `domain` - (Required) Verified domain name to generate DKIM tokens for.
* `mail_from_domain` - (Required) Subdomain (of above domain) which is to be used as MAIL FROM address (Required for DMARC validation)

## Followup Steps

Find out more about setting MAIL FROM domain with Amazon SES in [docs](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/mail-from-set.html).

* Create an MX record used to verify SES MAIL FROM setup

* (Optionally) If you want your emails to pass SPF checks, you must publish an SPF record to the DNS server of the custom MAIL FROM domain.

## Example Usage

```hcl
resource "aws_ses_domain_identity" "example" {
  domain = "example.com"
}

resource "aws_ses_domain_mail_from" "example" {
  domain           = "example.com"
  mail_from_domain = "bounce.example.com"
}

resource "aws_route53_record" "example_amazonses_mail_from_mx_record" {
  zone_id = "ABCDEFGHIJ123" # Change to appropriate Route53 Zone ID
  name    = "bounce.example.com"
  type    = "MX"
  ttl     = "600"
  records = ["10 feedback-smtp.us-east-1.amazonses.com"] # Change to the region in which `aws_ses_domain_identity.example` is created
}

# For SPF Check
resource "aws_route53_record" "example_amazonses_mail_from_mx_record" {
  zone_id = "ABCDEFGHIJ123" # Change to appropriate Route53 Zone ID
  name    = "bounce.example.com"
  type    = "TXT"
  ttl     = "600"
  records = ["v=spf1 include:amazonses.com -all"]
}

```

## Import

MAIL FROM domain can be imported using the `domain` attribute, e.g.

```
$ terraform import aws_ses_domain_mail_from.example example.com
```
