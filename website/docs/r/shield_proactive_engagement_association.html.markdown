---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_proactive_engagement_association"
description: |-
  Terraform resource for managing an AWS Shield Proactive Engagement Association.
---

# Resource: aws_shield_proactive_engagement_association

Initializes proactive engagement and sets the list of contacts for the Shield Response Team (SRT) to use. You must provide at least one phone number in the emergency contact list.

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
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"
}

resource "aws_shield_proactive_engagement_association" "test" {
  enabled = true
  emergency_contacts {
    contact_notes = "Notes"
    email_address = "test@company.com"
    phone_number = "+12358132134"
  }
  emergency_contacts {
    contact_notes = "Notes 2"
    email_address = "test2@company.com"
    phone_number = "+12358132134"
  }
  depends_on = [aws_shield_drt_access_role_arn_association.test]
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Boolean value indicating if Proactive Engagement should be enabled or nota
* `emergency_contacts` - (Required) One or more emergency contacts are required if Proactive Engagement is to be enabled. See [`emergency_contacts`](#emergency_contacts).


### emergency_contacts

* `email_address` - (Required) A valid email address that will be used for this contact.
* `phone_number` - (Optional) A phone number, starting with `+` and up to 15 digits that will be used for this contact.
* `contact_notes` - (Optional) Additional notes regarding the contact.

## Attribute Reference

This resource exports no additional attributes.
