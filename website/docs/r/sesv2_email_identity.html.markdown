---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity"
description: |-
  Provides an SES email identity resource
---

# Resource: aws_sesv2_email_identity

Provides an SES email identity resource

## Argument Reference

The following arguments are supported:

* `identity` - (Required) The email address or domain to assign to SES

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the email identity.
* `identity_type` - The email identity type.
* `origin` - A string that indicates how DKIM was configured for the domain
  identity.
* `dkim_tokens` - A list of strings that indicates which DNS records are used to
  validate the domain identity.

## Example Usage

```terraform
resource "aws_sesv2_email_identity" "example" {
  identity = "email@example.com"
}
```

## Import

SES email identities can be imported using the identity address or domain name.

```
$ terraform import aws_sesv2_email_identity.example email@example.com
```
