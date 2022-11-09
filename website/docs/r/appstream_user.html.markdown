---
subcategory: "AppStream 2.0"
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
  user_name           = "EMAIL"
  first_name          = "FIRST NAME"
  last_name           = "LAST NAME"
}
```

## Argument Reference

The following arguments are required:

* `authentication_type` - (Required) Authentication type for the user. You must specify USERPOOL. Valid values: `API`, `SAML`, `USERPOOL`
* `user_name` - (Required) Email address of the user.

The following arguments are optional:

* `enabled` - (Optional) Whether the user in the user pool is enabled.
* `first_name` - (Optional) First name, or given name, of the user.
* `last_name` - (Optional) Last name, or surname, of the user.
* `send_email_notification` - (Optional) Send an email notification.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the appstream user.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the user was created.
* `id` - Unique ID of the appstream user.
* `status` - Status of the user in the user pool.

## Import

`aws_appstream_user` can be imported using the `user_name` and `authentication_type` separated by a slash (`/`), e.g.,

```
$ terraform import aws_appstream_user.example UserName/AuthenticationType
```
