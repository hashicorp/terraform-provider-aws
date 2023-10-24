---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity_mail_from_attributes"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Mail From Attributes.
---

# Resource: aws_sesv2_email_identity_mail_from_attributes

Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Mail From Attributes.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}

resource "aws_sesv2_email_identity_mail_from_attributes" "example" {
  email_identity = aws_sesv2_email_identity.example.email_identity

  behavior_on_mx_failure = "REJECT_MESSAGE"
  mail_from_domain       = "subdomain.${aws_sesv2_email_identity.example.email_identity}"
}
```

## Argument Reference

This resource supports the following arguments:

* `email_identity` - (Required) The verified email identity.
* `behavior_on_mx_failure` - (Optional) The action to take if the required MX record isn't found when you send an email. Valid values: `USE_DEFAULT_VALUE`, `REJECT_MESSAGE`.
* `mail_from_domain` - (Optional) The custom MAIL FROM domain that you want the verified identity to use. Required if `behavior_on_mx_failure` is `REJECT_MESSAGE`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Email Identity Mail From Attributes using the `email_identity`. For example:

```terraform
import {
  to = aws_sesv2_email_identity_mail_from_attributes.example
  id = "example.com"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Email Identity Mail From Attributes using the `email_identity`. For example:

```console
% terraform import aws_sesv2_email_identity_mail_from_attributes.example example.com
```
