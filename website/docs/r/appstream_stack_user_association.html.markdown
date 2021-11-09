---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_stack_user_association"
description: |-
  Manages an AppStream Stack User association.
---

# Resource: aws_appstream_stack_user_association

Manages an AppStream Stack User association.

## Example Usage

```terraform
resource "aws_appstream_stack" "test" {
  name = "STACK NAME"
}

resource "aws_appstream_user" "test" {
  authentication_type = "USERPOOL"
  user_name           = "EMAIL"
}

resource "aws_appstream_stack_user_association" "test" {
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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique ID of the appstream stack user association.


## Import

AppStream Stack User Association can be imported by using the `stack_name` , `user_name` and `authentication_type` separated by a slash (`/`), e.g.,

```
$ terraform import aws_appstream_stack_user_association.example stackName/userName/auhtenticationType
```
