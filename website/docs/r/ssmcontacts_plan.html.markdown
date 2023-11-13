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
* `stage` - (Required) One or more configuration blocks for specifying a list of stages that the escalation plan or engagement plan uses to engage contacts and contact methods. See [Stage](#stage) below for more details.

### Stage

A stage specifies a set amount of time that an escalation plan or engagement plan engages the specified contacts or contact methods.

The `stage` block supports the following:

* `duration_in_minutes` - (Required) The time to wait until beginning the next stage. The duration can only be set to 0 if a target is specified.
* `target` - (Required) One or more configuration blocks for specifying the contacts or contact methods that the escalation plan or engagement plan is engaging. See [Target](#target) below for more details.

### Target

A target specifies the contact or contact channel that's being engaged.

The `target` block supports the following:

* `channel_target_info` - (Optional) A configuration block for specifying information about the contact channel that Incident Manager engages. See [Channel Target Info](#channel-target-info) for more details.
* `contact_target_info` - (Optional) A configuration block for specifying information about the contact that Incident Manager engages. See [Contact Target Info](#contact-target-info) for more details.

### Channel Target Info

Channel target info specifies information about the contact channel that Incident Manager uses to engage the contact.

The `channel_target_info` block supports the following:

* `contact_channel_id` - (Required) The Amazon Resource Name (ARN) of the contact channel.
* `retry_interval_in_minutes` - (Optional) The number of minutes to wait before retrying to send engagement if the engagement initially failed.

### Contact Target Info

Contact target info specifies the contact that Incident Manager is engaging during an incident.

The `contact_target_info` block supports the following:

* `contact_id` - (Optional) The Amazon Resource Name (ARN) of the contact.
* `is_essential` - (Optional) A Boolean value determining if the contact's acknowledgement stops the progress of stages in the plan.

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
