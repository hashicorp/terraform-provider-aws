---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_plan"
description: |-
  Terraform data source for managing an AWS SSM Contact Plan.
---

# Data Source: aws_ssmcontacts_plan

Terraform data source for managing a Plan of an AWS SSM Contact.

## Example Usage

### Basic Usage

```terraform
data "aws_ssmcontacts_plan" "test" {
  contact_id = "arn:aws:ssm-contacts:us-west-2:123456789012:contact/contactalias"
}
```

## Argument Reference

The following arguments are required:

* `contact_id` - (Required) The Amazon Resource Name (ARN) of the contact or escalation plan.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `stage` - List of stages. A contact has an engagement plan with stages that contact specified contact channels. An escalation plan uses stages that contact specified contacts.
