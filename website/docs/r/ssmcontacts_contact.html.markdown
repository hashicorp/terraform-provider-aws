---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_contact"
description: |-
  Terraform resource for managing an AWS SSM Contact.
---

# Resource: aws_ssmcontacts_contact

Terraform resource for managing an AWS SSM Contact.

~> **NOTE:** A contact implicitly depends on a replication set. If you configured your replication set in Terraform, we recommend you add it to the `depends_on` argument for the Terraform Contact Resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmcontacts_contact" "example" {
  alias = "alias"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

### Usage With All Fields

```terraform
resource "aws_ssmcontacts_contact" "example" {
  alias        = "alias"
  display_name = "displayName"
  type         = "ESCALATION"

  tags = {
    key = "value"
  }

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

### On-call Schedule Usage

```terraform
resource "aws_ssmcontacts_contact" "oncall_contact" {
  alias = "oncall-contact"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.example]
}

resource "aws_ssmcontacts_rotation" "example" {
  contact_ids = [aws_ssmcontacts_contact.oncall_contact.arn]
  name        = "example-rotation"

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 0
    }
  }

  time_zone_id = "America/Los_Angeles"

  depends_on = [aws_ssmincidents_replication_set.example]
}

resource "aws_ssmcontacts_contact" "example" {
  alias        = "oncall-schedule"
  display_name = "Example On-call Schedule"
  type         = "ONCALL_SCHEDULE"
  rotation_ids = [aws_ssmcontacts_rotation.example.arn]

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

## Argument Reference

The following arguments are required:

- `alias` - (Required) A unique and identifiable alias for the contact or escalation plan. Must be between 1 and 255 characters, and may contain alphanumerics, underscores (`_`), and hyphens (`-`).
- `type` - (Required) The type of contact engaged. A single contact is type `PERSONAL`, an escalation plan is type `ESCALATION`, and an on-call schedule is type `ONCALL_SCHEDULE`.

The following arguments are optional:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `display_name` - (Optional) Full friendly name of the contact or escalation plan. If set, must be between 1 and 255 characters, and may contain alphanumerics, underscores (`_`), hyphens (`-`), periods (`.`), and spaces.
- `rotation_ids` - (Optional) List of rotation IDs associated with the contact. Required when `type` is `ONCALL_SCHEDULE`.
- `tags` - (Optional) Key-value tags for the monitor. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The Amazon Resource Name (ARN) of the contact or escalation plan.
- `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Contact using the `ARN`. For example:

```terraform
import {
  to = aws_ssmcontacts_contact.example
  id = "{ARNValue}"
}
```

Using `terraform import`, import SSM Contact using the `ARN`. For example:

```console
% terraform import aws_ssmcontacts_contact.example {ARNValue}
```
