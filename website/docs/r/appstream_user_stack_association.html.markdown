---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_user_stack_association"
description: |-
  Manages an AppStream User Stack association.
---

# Resource: aws_appstream_user_stack_association

Manages an AppStream User Stack association.

## Example Usage

```terraform
resource "aws_appstream_stack" "test" {
  name = "STACK NAME"
}

resource "aws_appstream_user" "test" {
  authentication_type = "USERPOOL"
  user_name           = "EMAIL"
}

resource "aws_appstream_user_stack_association" "test" {
  authentication_type = aws_appstream_user.test.authentication_type
  stack_name          = aws_appstream_stack.test.name
  user_name           = aws_appstream_user.test.user_name
}
```

## Argument Reference

The following arguments are required:

* `authentication_type` - (Required) Authentication type for the user.
* `stack_name` (Required) Name of the stack that is associated with the user.
* `user_name` (Required) Email address of the user who is associated with the stack.

The following arguments are optional:

* `send_email_notification` - (Optional) Whether a welcome email is sent to a user after the user is created in the user pool.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the appstream User Stack association.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppStream User Stack Association using the `user_name`, `authentication_type`, and `stack_name`, separated by a slash (`/`). For example:

```terraform
import {
  to = aws_appstream_user_stack_association.example
  id = "userName/auhtenticationType/stackName"
}
```

Using `terraform import`, import AppStream User Stack Association using the `user_name`, `authentication_type`, and `stack_name`, separated by a slash (`/`). For example:

```console
% terraform import aws_appstream_user_stack_association.example userName/auhtenticationType/stackName
```
