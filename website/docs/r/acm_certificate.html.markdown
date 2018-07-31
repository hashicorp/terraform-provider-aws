---
layout: "aws"
page_title: "AWS: aws_acm_certificate"
sidebar_current: "docs-aws-resource-acm-certificate"
description: |-
  Requests and manages a certificate from Amazon Certificate Manager (ACM).
---

# aws_acm_certificate

The ACM certificate resource allows requesting and management of certificates
from the Amazon Certificate Manager.

It deals with requesting certificates and managing their attributes and life-cycle.
This resource does not deal with validation of a certificate but can provide inputs
for other resources implementing the validation. It does not wait for a certificate to be issued.
Use a [`aws_acm_certificate_validation`](acm_certificate_validation.html) resource for this.

Most commonly, this resource is used to together with [`aws_route53_record`](route53_record.html) and
[`aws_acm_certificate_validation`](acm_certificate_validation.html) to request a DNS validated certificate,
deploy the required validation records and wait for validation to complete.

Domain validation through E-Mail is also supported but should be avoided as it requires a manual step outside
of Terraform.

## Example Usage

```hcl
resource "aws_acm_certificate" "cert" {
  domain_name = "example.com"
  validation_method = "DNS"
  tags {
    Environment = "test"
  }
}

#example with subject_alternative_names and domain_validation_options
resource "aws_acm_certificate" "cert" {
  domain_name               = "yolo.example.io"
  validation_method         = "EMAIL"
  subject_alternative_names = ["app1.yolo.example.io", "yolo.example.io"]

  domain_validation_options = [
    {
      domain_name       = "yolo.example.io"
      validation_domain = "example.io"
    },
    {
      domain_name       = "app1.yolo.example.io"
      validation_domain = "example.io"
    },
  ]
}

#basic example
resource "aws_acm_certificate" "cert" {
  domain_name               = "yolo.example.io"
  validation_method         = "EMAIL"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) A fully qualified domain name (FQDN) in the certificate. For example, www.example.com or example.com .
* `subject_alternative_names` - (Optional) One or more domain names (subject alternative names) included in the certificate. This list contains the domain names that are bound to the public key that is contained in the certificate. The subject alternative names include the canonical domain name (CN) of the certificate and additional domain names that can be used to connect to the website.
* `validation_method` - (Required) Which method to use for validation. `DNS` or `EMAIL` are valid, `NONE` can be used for certificates that were imported into ACM and then into Terraform.
* `domain_validaton_options` - (Optional) Contains information about the initial validation of each domain name that occurs. This is an array of maps that contains information about which validation_domain to use for domains in the subject_alternative_names list.

* `tags` - (Optional) A mapping of tags to assign to the resource.

Domain Validation Options objects accept the following attributes

* `domain_name` - (Required) A fully qualified domain name (FQDN) in the certificate. For example, www.example.com or example.com .
* `validation_domain` - (Required) The domain name that ACM used to send domain validation emails

## Attributes Reference

The following additional attributes are exported:

* `id` - The ARN of the certificate
* `arn` - The ARN of the certificate
* `certificate_details` - A list of attributes to feed into other resources to complete certificate validation. Can have more than one element, e.g. if SANs are defined. 

Certificate Details objects export the following attributes:

* `domain_name` - A fully qualified domain name (FQDN) in the certificate. For example, www.example.com or example.com .
* `resource_record_name` - The name of the DNS record to create in your domain. This is supplied by ACM.
* `resource_record_type` - The type of DNS record. Currently this can be CNAME .
* `resource_record_value` - The value of the CNAME record to add to your DNS database. This is supplied by ACM. 
* `validation_domain` - The domain name that ACM used to send domain validation emails.
* `validation_method` - One of EMAIl or DNS
* `validation_emails` - A list of email addresses that ACM used to send domain validation emails.

## Import

Certificates can be imported using their ARN, e.g.

```
$ terraform import aws_acm_certificate.cert arn:aws:acm:eu-central-1:123456789012:certificate/7e7a28d2-163f-4b8f-b9cd-822f96c08d6a
```

~> **WARNING:** Importing certificates that are not `AMAZON_ISSUED` is supported but may lead to fragile terraform projects: Once such a resource is destroyed, it can't be recreated.
