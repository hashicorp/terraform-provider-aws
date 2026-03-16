---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_user"
description: |-
  Manages an AWS WorkMail User.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_workmail_user

Manages an AWS WorkMail User.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_organization" "example" {
  organization_alias = "example-org"
  delete_directory   = true
}

resource "aws_workmail_user" "example" {
  organization_id = aws_workmail_organization.example.organization_id
  name            = "example-user"
  display_name    = "Example User"
  password        = "ExamplePassword123!"
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Display name of the user.
* `name` - (Required) Username of the user.
* `organization_id` - (Required) Identifier of the WorkMail organization where the user is managed.

The following arguments are optional:

* `city` - (Optional) City where the user is located.
* `company` - (Optional) Company associated with the user.
* `country` - (Optional) Country where the user is located.
* `department` - (Optional) Department associated with the user.
* `first_name` - (Optional) First name of the user.
* `hidden_from_global_address_list` - (Optional) Whether to hide the user from the global address list. Defaults to `false`.
* `identity_provider_user_id` - (Optional) User ID from IAM Identity Center associated with the user.
* `initials` - (Optional) Initials of the user.
* `job_title` - (Optional) Job title of the user.
* `last_name` - (Optional) Last name of the user.
* `office` - (Optional) Office where the user is located.
* `password` - (Optional, Sensitive) Password used when creating the user. Changing this value forces replacement.
* `role` - (Optional) Role of the user. Defaults to `USER`.
* `street` - (Optional) Street address of the user.
* `telephone` - (Optional) Telephone number of the user.
* `zip_code` - (Optional) ZIP or postal code of the user.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `disabled_date` - Timestamp when the user was disabled from WorkMail use.
* `email` - Primary email address of the user.
* `enabled_date` - Timestamp when the user was enabled for WorkMail use.
* `identity_provider_identity_store_id` - Identity store ID from IAM Identity Center associated with the user.
* `mailbox_deprovisioned_date` - Timestamp when the mailbox was removed for the user.
* `mailbox_provisioned_date` - Timestamp when the mailbox was created for the user.
* `state` - Current WorkMail state of the user.
* `user_id` - Identifier of the user.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

This resource does not support configurable timeouts.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_workmail_user.example
  identity = {
    organization_id = "m-1234567890abcdef1234567890abcdef"
    user_id         = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_workmail_user" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required
* `organization_id` - Identifier of the WorkMail organization.
* `user_id` - Identifier of the user.

#### Optional
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkMail User using the resource ID. For example:

```terraform
import {
  to = aws_workmail_user.example
  id = "m-1234567890abcdef1234567890abcdef,12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import WorkMail User using `organization_id,user_id`. For example:

```console
% terraform import aws_workmail_user.example m-1234567890abcdef1234567890abcdef,12345678-1234-1234-1234-123456789012
```
