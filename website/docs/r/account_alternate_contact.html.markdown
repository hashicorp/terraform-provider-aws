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

This resource supports the following arguments:

* `account_id` - (Optional) ID of the target account when managing member accounts. Will manage current user's account by default if omitted.
* `alternate_contact_type` - (Required) Type of the alternate contact. Allowed values are: `BILLING`, `OPERATIONS`, `SECURITY`.
* `email_address` - (Required) An email address for the alternate contact.
* `name` - (Required) Name of the alternate contact.
* `phone_number` - (Required) Phone number for the alternate contact.
* `title` - (Required) Title for the alternate contact.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Alternate Contact for the current or another account using the `alternate_contact_type`. For example:

Import the Alternate Contact for the current account:

```terraform
import {
  to = aws_account_alternate_contact.operations
  id = "OPERATIONS"
}
```

Import the Alternate Contact for another account using the `account_id` and `alternate_contact_type` separated by a forward slash (`/`):

```terraform
import {
  to = aws_account_alternate_contact.operations
  id = "1234567890/OPERATIONS"
}
```

**Using `terraform import` to import** the Alternate Contact for the current or another account using the `alternate_contact_type`. For example:

Import the Alternate Contact for the current account:

```console
% terraform import aws_account_alternate_contact.operations OPERATIONS
```

Import the Alternate Contact for another account using the `account_id` and `alternate_contact_type` separated by a forward slash (`/`):

```console
% terraform import aws_account_alternate_contact.operations 1234567890/OPERATIONS
```
