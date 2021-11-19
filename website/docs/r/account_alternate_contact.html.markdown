---
subcategory: "Account"
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

* `alternate_contact_type` - (Required) The type of the alternate contact. Allowed values are: `BILLING`, `OPERATIONS`, `SECURITY`.
* `email_address` - (Required) An email address for the alternate contact.
* `name` - (Required) The name of the alternate contact.
* `phone_number` - (Required) A phone number for the alternate contact.
* `title` - (Required) A title for the alternate contact.

## Attributes Reference

No additional attributes are exported.

## Import

The current Alternate Contact can be imported using the `alternate_contact_type`, e.g.,

```
$ terraform import aws_account_alternate_contact.operations OPERATIONS
```
