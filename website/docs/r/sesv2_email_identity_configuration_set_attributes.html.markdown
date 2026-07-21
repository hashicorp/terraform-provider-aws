---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity_configuration_set_attributes"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Email Identity Configuration Set association.
---

# Resource: aws_sesv2_email_identity_configuration_set_attributes

Manages the configuration set association for an SESv2 email identity.

Use this resource instead of the `configuration_set_name` argument on `aws_sesv2_email_identity` when you need to break a circular dependency: for example, when an `aws_sesv2_configuration_set` with click/open tracking requires the domain identity to already exist, and `aws_sesv2_email_identity` requires the configuration set to already exist.

~> **Note:** Do not use `configuration_set_name` in `aws_sesv2_email_identity` and `aws_sesv2_email_identity_configuration_set_attributes` on the same identity at the same time. Destroying this resource will disassociate the configuration set (set it to empty), not restore any previously configured value.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}

resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_email_identity_configuration_set_attributes" "example" {
  email_identity         = aws_sesv2_email_identity.example.email_identity
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
}
```

### Breaking a Circular Dependency

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}

resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"

  tracking_options {
    custom_redirect_domain = "click.example.com"
  }
}

resource "aws_sesv2_email_identity_configuration_set_attributes" "example" {
  email_identity         = aws_sesv2_email_identity.example.email_identity
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `email_identity` - (Required, Forces new resource) The verified email identity (domain or email address).
* `configuration_set_name` - (Required) Name of the configuration set to associate with the email identity.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Email Identity Configuration Set Attributes using the `email_identity`. For example:

```terraform
import {
  to = aws_sesv2_email_identity_configuration_set_attributes.example
  id = "example.com"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Email Identity Configuration Set Attributes using the `email_identity`. For example:

```console
% terraform import aws_sesv2_email_identity_configuration_set_attributes.example example.com
```
