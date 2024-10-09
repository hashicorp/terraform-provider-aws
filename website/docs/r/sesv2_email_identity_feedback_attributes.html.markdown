---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity_feedback_attributes"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Feedback Attributes.
---

# Resource: aws_sesv2_email_identity_feedback_attributes

Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Feedback Attributes.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}

resource "aws_sesv2_email_identity_feedback_attributes" "example" {
  email_identity           = aws_sesv2_email_identity.example.email_identity
  email_forwarding_enabled = true
}
```

## Argument Reference

This resource supports the following arguments:

* `email_identity` - (Required) The email identity.
* `email_forwarding_enabled` - (Optional) Sets the feedback forwarding configuration for the identity.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Email Identity Feedback Attributes using the `email_identity`. For example:

```terraform
import {
  to = aws_sesv2_email_identity_feedback_attributes.example
  id = "example.com"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Email Identity Feedback Attributes using the `email_identity`. For example:

```console
% terraform import aws_sesv2_email_identity_feedback_attributes.example example.com
```
