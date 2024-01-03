---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_contact"
description: |-
  Terraform data source for managing an AWS SSM Contact.
---

# Data Source: aws_ssmcontacts_contact

Terraform data source for managing an AWS SSM Contact.

## Example Usage

### Basic Usage

```terraform
data "aws_ssmcontacts_contact" "example" {
  arn = "arn:aws:ssm-contacts:us-west-2:123456789012:contact/contactalias"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) The Amazon Resource Name (ARN) of the contact or escalation plan.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `alias` - A unique and identifiable alias of the contact or escalation plan.
* `type` - The type of contact engaged. A single contact is type `PERSONAL` and an escalation plan is type `ESCALATION`.
* `display_name` - Full friendly name of the contact or escalation plan.
* `tags` - Map of tags to assign to the resource.
