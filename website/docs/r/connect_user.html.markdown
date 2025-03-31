---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_user"
description: |-
  Provides details about a specific Amazon Connect User
---

# Resource: aws_connect_user

Provides an Amazon Connect User resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

### Basic

```terraform
resource "aws_connect_user" "example" {
  instance_id        = aws_connect_instance.example.id
  name               = "example"
  password           = "Password123"
  routing_profile_id = aws_connect_routing_profile.example.routing_profile_id

  security_profile_ids = [
    aws_connect_security_profile.example.security_profile_id
  ]

  identity_info {
    first_name = "example"
    last_name  = "example2"
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
```

### With hierarchy_group_id

```terraform
resource "aws_connect_user" "example" {
  instance_id        = aws_connect_instance.example.id
  name               = "example"
  password           = "Password123"
  routing_profile_id = aws_connect_routing_profile.example.routing_profile_id
  hierarchy_group_id = aws_connect_user_hierarchy_group.example.hierarchy_group_id

  security_profile_ids = [
    aws_connect_security_profile.example.security_profile_id
  ]

  identity_info {
    first_name = "example"
    last_name  = "example2"
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
```

### With identity_info filled

```terraform
resource "aws_connect_user" "example" {
  instance_id        = aws_connect_instance.example.id
  name               = "example"
  password           = "Password123"
  routing_profile_id = aws_connect_routing_profile.example.routing_profile_id

  security_profile_ids = [
    aws_connect_security_profile.example.security_profile_id
  ]

  identity_info {
    email           = "example@example.com"
    first_name      = "example"
    last_name       = "example2"
    secondary_email = "secondary@example.com"
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
```

### With phone_config phone type as desk phone

```terraform
resource "aws_connect_user" "example" {
  instance_id        = aws_connect_instance.example.id
  name               = "example"
  password           = "Password123"
  routing_profile_id = aws_connect_routing_profile.example.routing_profile_id

  security_profile_ids = [
    aws_connect_security_profile.example.security_profile_id
  ]

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
```

### With multiple Security profile ids specified in security_profile_ids

```terraform
resource "aws_connect_user" "example" {
  instance_id        = aws_connect_instance.example.id
  name               = "example"
  password           = "Password123"
  routing_profile_id = aws_connect_routing_profile.example.routing_profile_id

  security_profile_ids = [
    aws_connect_security_profile.example.security_profile_id,
    aws_connect_security_profile.example2.security_profile_id,
  ]

  phone_config {
    after_contact_work_time_limit = 0
    auto_accept                   = false
    desk_phone_number             = "+112345678912"
    phone_type                    = "DESK_PHONE"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `directory_user_id` - (Optional) The identifier of the user account in the directory used for identity management. If Amazon Connect cannot access the directory, you can specify this identifier to authenticate users. If you include the identifier, we assume that Amazon Connect cannot access the directory. Otherwise, the identity information is used to authenticate users from your directory. This parameter is required if you are using an existing directory for identity management in Amazon Connect when Amazon Connect cannot access your directory to authenticate users. If you are using SAML for identity management and include this parameter, an error is returned.
* `hierarchy_group_id` - (Optional) The identifier of the hierarchy group for the user.
* `identity_info` - (Optional) A block that contains information about the identity of the user. Documented below.
* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `name` - (Required) The user name for the account. For instances not using SAML for identity management, the user name can include up to 20 characters. If you are using SAML for identity management, the user name can include up to 64 characters from `[a-zA-Z0-9_-.\@]+`.
* `password` - (Optional) The password for the user account. A password is required if you are using Amazon Connect for identity management. Otherwise, it is an error to include a password.
* `phone_config` - (Required) A block that contains information about the phone settings for the user. Documented below.
* `routing_profile_id` - (Required) The identifier of the routing profile for the user.
* `security_profile_ids` - (Required) A list of identifiers for the security profiles for the user. Specify a minimum of 1 and maximum of 10 security profile ids. For more information, see [Best Practices for Security Profiles](https://docs.aws.amazon.com/connect/latest/adminguide/security-profile-best-practices.html) in the Amazon Connect Administrator Guide.
* `tags` - (Optional) Tags to apply to the user. If configured with a provider
[`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `identity_info` block supports the following arguments:

* `email` - (Optional) The email address. If you are using SAML for identity management and include this parameter, an error is returned. Note that updates to the `email` is supported. From the [UpdateUserIdentityInfo API documentation](https://docs.aws.amazon.com/connect/latest/APIReference/API_UpdateUserIdentityInfo.html) it is strongly recommended to limit who has the ability to invoke `UpdateUserIdentityInfo`. Someone with that ability can change the login credentials of other users by changing their email address. This poses a security risk to your organization. They can change the email address of a user to the attacker's email address, and then reset the password through email. For more information, see [Best Practices for Security Profiles](https://docs.aws.amazon.com/connect/latest/adminguide/security-profile-best-practices.html) in the Amazon Connect Administrator Guide.
* `first_name` - (Optional) The first name. This is required if you are using Amazon Connect or SAML for identity management. Minimum length of 1. Maximum length of 100.
* `last_name` - (Optional) The last name. This is required if you are using Amazon Connect or SAML for identity management. Minimum length of 1. Maximum length of 100.
* `secondary_email` - (Optional) The secondary email address. If present, email notifications will be sent to this email address instead of the primary one.

A `phone_config` block supports the following arguments:

* `after_contact_work_time_limit` - (Optional) The After Call Work (ACW) timeout setting, in seconds. Minimum value of 0.
* `auto_accept` - (Optional) When Auto-Accept Call is enabled for an available agent, the agent connects to contacts automatically.
* `desk_phone_number` - (Optional) The phone number for the user's desk phone. Required if `phone_type` is set as `DESK_PHONE`.
* `phone_type` - (Required) The phone type. Valid values are `DESK_PHONE` and `SOFT_PHONE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the user.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the user
separated by a colon (`:`).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `user_id` - The identifier for the user.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Users using the `instance_id` and `user_id` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_connect_user.example
  id = "f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5"
}
```

Using `terraform import`, import Amazon Connect Users using the `instance_id` and `user_id` separated by a colon (`:`). For example:

```console
% terraform import aws_connect_user.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5
```
