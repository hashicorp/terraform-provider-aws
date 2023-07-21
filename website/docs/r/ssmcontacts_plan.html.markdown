---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_plan"
description: |-
  Terraform resource for managing an AWS SSM Contact Plan.
---

# Resource: aws_ssmcontacts_plan

Terraform resource for managing an AWS SSM Contact Plan.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmcontacts_plan" "example" {
  contact_id = "arn:aws:ssm-contacts:us-west-2:123456789012:contact/contactalias"
  stage {
    duration_in_minutes = 1
  }
}
```

### Usage with SSM Contact

```terraform
resource "aws_ssmcontacts_contact" "contact" {
  alias = "alias"
  type  = "PERSONAL"
}

resource "aws_ssmcontacts_plan" "plan" {
  contact_id = aws_ssmcontacts_contact.contact.arn
  stage {
    duration_in_minutes = 1
  }
}
```

### Usage With All Fields

```terraform
resource "aws_ssmcontacts_contact" "escalation_plan" {
  alias = "escalation-plan-alias"
  type  = "ESCALATION"
}

resource "aws_ssmcontacts_contact" "contact_one" {
  alias = "alias"
  type  = "PERSONAL"
}

resource "aws_ssmcontacts_contact" "contact_two" {
  alias = "alias"
  type  = "PERSONAL"
}

resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.escalation_plan.arn

  stage {
    duration_in_minutes = 0

    target {
      contact_target_info {
        is_essential = false
        contact_id   = aws_ssmcontacts_contact.contact_one.arn
      }
    }

    target {
      contact_target_info {
        is_essential = true
        contact_id   = aws_ssmcontacts_contact.contact_two.arn
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `contact_id` - (Required) The Amazon Resource Name (ARN) of the contact or escalation plan.
* `stage` - (Required) List of stages. A contact has an engagement plan with stages that contact specified contact channels. An escalation plan uses stages that contact specified contacts.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Contact Plan using the Contact ARN. For example:

```terraform
import {
  to = aws_ssmcontacts_plan.example
  id = "{ARNValue}"
}
```

Using `terraform import`, import SSM Contact Plan using the Contact ARN. For example:

```console
% terraform import aws_ssmcontacts_plan.example {ARNValue}
```
