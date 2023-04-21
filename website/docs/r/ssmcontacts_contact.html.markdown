---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_contact"
description: |-
  Terraform resource for managing an AWS SSM Contact.
---

# Resource: aws_ssmcontacts_contact

Terraform resource for managing an AWS SSM Contact.

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

~> **NOTE:** A contact implicitly depends on a replication set. If you configured your replication set in Terraform, we recommend you add it to the `depends_on` argument for the Terraform Contact Resource.

The following arguments are required:

- `alias` - (Required) A unique and identifiable alias for the contact or escalation plan.

- `type` - (Required) The type of contact engaged. A single contact is type PERSONAL and an escalation
  plan is type ESCALATION.

The following arguments are optional:

- `display_name` - (Optional) Full friendly name of the contact or escalation plan.

- `tags` - (Optional) Map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - The Amazon Resource Name (ARN) of the contact or escalation plan.

- `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Import SSM Contact using the `ARN`. For example:

```
$ terraform import aws_ssmcontacts_contact.example {ARNValue}
```
