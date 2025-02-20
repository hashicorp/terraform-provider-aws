---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_in_group"
description: |-
  Adds the specified user to the specified group.
---

# Resource: aws_cognito_user_in_group

Adds the specified user to the specified group.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example"

  password_policy {
    temporary_password_validity_days = 7
    minimum_length                   = 6
    require_uppercase                = false
    require_symbols                  = false
    require_numbers                  = false
  }
}

resource "aws_cognito_user" "example" {
  user_pool_id = aws_cognito_user_pool.example.id
  username     = "example"
}

resource "aws_cognito_user_group" "example" {
  user_pool_id = aws_cognito_user_pool.example.id
  name         = "example"
}

resource "aws_cognito_user_in_group" "example" {
  user_pool_id = aws_cognito_user_pool.example.id
  group_name   = aws_cognito_user_group.example.name
  username     = aws_cognito_user.example.username
}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) The user pool ID of the user and group.
* `group_name` - (Required) The name of the group to which the user is to be added.
* `username` - (Required) The username of the user to be added to the group.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito User Groups using the `user_pool_id`/`group_name`/`username` attributes concatenated. For example:

```terraform
import {
  to = aws_cognito_user_in_group.example
  id = "us-east-1_vG78M4goG/group-name/user-name"
}
```

Using `terraform import`, import Cognito User in Groups using the `user_pool_id` and `group_name` and `username` separated by a forward slash (`/`). For example:

```console
% terraform import aws_cognito_user_in_group.example us-east-1_vG78M4goG/group-name/user-name
```
