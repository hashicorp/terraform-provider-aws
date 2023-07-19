---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_email_identity"
description: |-
  Provides an SES email identity resource
---

# Resource: aws_ses_email_identity

Provides an SES email identity resource

## Argument Reference

This resource supports the following arguments:

* `email` - (Required) The email address to assign to SES.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the email identity.

## Example Usage

```terraform
resource "aws_ses_email_identity" "example" {
  email = "email@example.com"
}
```

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES email identities using the email address. For example:

```terraform
import {
  to = aws_ses_email_identity.example
  id = "email@example.com"
}
```

Using `terraform import`, import SES email identities using the email address. For example:

```console
% terraform import aws_ses_email_identity.example email@example.com
```
