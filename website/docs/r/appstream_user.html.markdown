---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_user"
description: |-
  Provides an AppStream user
---

# Resource: aws_appstream_user

Provides an AppStream user.

## Example Usage

```terraform
resource "aws_appstream_user" "example" {
  authentication_type = "USERPOOL"
  user_name           = "EMAIL ADDRESS"
  first_name          = "FIRST NAME"
  last_name           = "LAST NAME"
  message_action      = "RESEND"
}
```

## Argument Reference

The following arguments are required:

* `authentication_type` - (Required) Authentication type for the user. You must specify USERPOOL. Valid values: `API`, `SAML`, `USERPOOL`
* `user_name` - (Required) Email address of the user.

The following arguments are optional:

* `first_name` - (Optional) First name, or given name, of the user.
* `last_name` - (Optional) Last name, or surname, of the user.
* `message_action` - (Optional) Action to take for the welcome email that is sent to a user after the user is created in the user pool. If you specify `SUPPRESS`, no email is sent. If you specify `RESEND`, do not specify the `first_name` or `last_name` of the user. If the value is null, the email is sent.


## Import

`aws_appstream_user` can be imported using the `user_name` and `authentication_type` separated by a slash (`/`), e.g.,

```
$ terraform import aws_appstream_user.example UserName/AuthenticationType
```
