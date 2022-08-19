---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_alternate_contact"
description: |-
  Manages the specified alternate contact attached to an AWS Account.
---

# Resource: aws_account_alternate_contact

Manages the specified alternate contact attached to an AWS Account.

## Example Usage

```terraform
resource "aws_account_alternate_contact" "operations" {

  alternate_contact_type = "OPERATIONS"

  name          = "Example"
  title         = "Example"
  email_address = "test@example.com"
  phone_number  = "+1234567890"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The ID of the target account when managing member accounts. Will manage current user's account by default if omitted.
* `alternate_contact_type` - (Required) The type of the alternate contact. Allowed values are: `BILLING`, `OPERATIONS`, `SECURITY`.
* `email_address` - (Required) An email address for the alternate contact.
* `name` - (Required) The name of the alternate contact.
* `phone_number` - (Required) A phone number for the alternate contact.
* `title` - (Required) A title for the alternate contact.

## Attributes Reference

No additional attributes are exported.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

The Alternate Contact for the current account can be imported using the `alternate_contact_type`, e.g.,

```
$ terraform import aws_account_alternate_contact.operations OPERATIONS
```

If you provide an account ID, the Alternate Contact can be imported using the `account_id` and `alternate_contact_type` separated by a forward slash (`/`) e.g.,

```
$ terraform import aws_account_alternate_contact.operations 1234567890/OPERATIONS
```
