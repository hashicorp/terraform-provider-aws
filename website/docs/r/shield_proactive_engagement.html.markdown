---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_proactive_engagement"
description: |-
  Terraform resource for managing a AWS Shield Proactive Engagement.
---

# Resource: aws_shield_proactive_engagement

Terraform resource for managing a AWS Shield Proactive Engagement.
Proactive engagement authorizes the Shield Response Team (SRT) to use email and phone to notify contacts about escalations to the SRT and to initiate proactive customer support.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = var.aws_shield_drt_access_role_arn
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        "Sid" : "",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "drt.shield.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_access_role_arn_association" "example" {
  role_arn = aws_iam_role.example.arn
}

resource "aws_shield_protection_group" "test" {
  protection_group_id = "example"
  aggregation         = "MAX"
  pattern             = "ALL"
}

resource "aws_shield_proactive_engagement" "test" {
  enabled = true

  emergency_contact {
    contact_notes = "Notes"
    email_address = "test@company.com"
    phone_number  = "+12358132134"
  }

  emergency_contact {
    contact_notes = "Notes 2"
    email_address = "test2@company.com"
    phone_number  = "+12358132134"
  }

  depends_on = [aws_shield_drt_access_role_arn_association.test]
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Boolean value indicating if Proactive Engagement should be enabled or not.
* `emergency_contact` - (Required) One or more emergency contacts. You must provide at least one phone number in the emergency contact list. See [`emergency_contacts`](#emergency_contacts).

### emergency_contacts

* `contact_notes` - (Optional) Additional notes regarding the contact.
* `email_address` - (Required) A valid email address that will be used for this contact.
* `phone_number` - (Optional) A phone number, starting with `+` and up to 15 digits that will be used for this contact.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield proactive engagement using the AWS account ID. For example:

```terraform
import {
  to = aws_shield_proactive_engagement.example
  id = "123456789012"
}
```

Using `terraform import`, import Shield proactive engagement using the AWS account ID. For example:

```console
% terraform import aws_shield_proactive_engagement.example 123456789012
```
