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

The following arguments are supported:

* `email_identity` - (Required) The verified email identity.
* `behavior_on_mx_failure` - (Optional) The action to take if the required MX record isn't found when you send an email. Valid values: `USE_DEFAULT_VALUE`, `REJECT_MESSAGE`.
* `mail_from_domain` - (Optional) The custom MAIL FROM domain that you want the verified identity to use. Required if `behavior_on_mx_failure` is `REJECT_MESSAGE`.

## Attributes Reference

No additional attributes are exported.

## Import

SESv2 (Simple Email V2) Email Identity Mail From Attributes can be imported using the `email_identity`, e.g.,

```
$ terraform import aws_sesv2_email_identity_mail_from_attributes.example example.com
```
