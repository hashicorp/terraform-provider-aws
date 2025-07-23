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

## Argument Reference

The following arguments are required:

- `alias` - (Required) A unique and identifiable alias for the contact or escalation plan. Must be between 1 and 255 characters, and may contain alphanumerics, underscores (`_`), and hyphens (`-`).
- `type` - (Required) The type of contact engaged. A single contact is type PERSONAL and an escalation
  plan is type ESCALATION.

The following arguments are optional:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `display_name` - (Optional) Full friendly name of the contact or escalation plan. If set, must be between 1 and 255 characters, and may contain alphanumerics, underscores (`_`), hyphens (`-`), periods (`.`), and spaces.
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
