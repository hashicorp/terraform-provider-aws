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

This resource supports the following arguments:

* `certificate_arn` - (Required) ARN of the certificate. This can be an Certificate Manager (ACM) certificate or an Identity and Access Management (IAM) server certificate. The certifiacte must reside in us-east-1.
* `description` - (Optional)  A description of the Domain Name.
* `domain_name` - (Required) Domain name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Appsync Domain Name.
* `appsync_domain_name` - Domain name that AppSync provides.
* `hosted_zone_id` - ID of your Amazon Route 53 hosted zone.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_domain_name` using the AppSync domain name. For example:

```terraform
import {
  to = aws_appsync_domain_name.example
  id = "example.com"
}
```

Using `terraform import`, import `aws_appsync_domain_name` using the AppSync domain name. For example:

```console
% terraform import aws_appsync_domain_name.example example.com
```
