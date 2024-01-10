---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity_mail_from_attributes"
description: |-
  Terraform data source for managing an AWS SESv2 (Simple Email V2) Email Identity Mail From Attributes.
---

# Data Source: aws_sesv2_email_identity_mail_from_attributes

Terraform data source for managing an AWS SESv2 (Simple Email V2) Email Identity Mail From Attributes.

## Example Usage

### Basic Usage

```terraform
data "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}

data "aws_sesv2_email_identity_mail_from_attributes" "example" {
  email_identity = data.aws_sesv2_email_identity.example.email_identity
}
```

## Argument Reference

The following arguments are required:

* `email_identity` - (Required) The name of the email identity.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `behavior_on_mx_failure` - The action to take if the required MX record isn't found when you send an email. Valid values: `USE_DEFAULT_VALUE`, `REJECT_MESSAGE`.
* `mail_from_domain` - The custom MAIL FROM domain that you want the verified identity to use.
