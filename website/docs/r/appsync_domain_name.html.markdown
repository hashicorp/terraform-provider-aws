---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_domain_name"
description: |-
  Provides an AppSync Domain Name.
---

# Resource: aws_appsync_domain_name

Provides an AppSync Domain Name.

## Example Usage

```terraform
resource "aws_appsync_domain_name" "example" {
  domain_name     = "api.example.com"
  certificate_arn = aws_acm_certificate.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `certificate_arn` - (Required) ARN of the certificate. This can be an Certificate Manager (ACM) certificate or an Identity and Access Management (IAM) server certificate. The certifiacte must reside in us-east-1.
* `description` - (Optional)  A description of the Domain Name.
* `domain_name` - (Required) Domain name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Appsync Domain Name.
* `appsync_domain_name` - Domain name that AppSync provides.
* `hosted_zone_id` - ID of your Amazon Route 53 hosted zone.

## Import

`aws_appsync_domain_name` can be imported using the AppSync domain name, e.g.,

```
$ terraform import aws_appsync_domain_name.example example.com
```
