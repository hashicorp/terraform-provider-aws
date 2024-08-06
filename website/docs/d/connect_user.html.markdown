---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_user"
description: |-
  Provides details about a specific Amazon Connect User.
---

# Data Source: aws_connect_user

Provides details about a specific Amazon Connect User.

## Example Usage

By `name`

```hcl
data "aws_connect_user" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `user_id`

```hcl
data "aws_connect_user" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  user_id     = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `user_id` is required.

This data source supports the following arguments:

* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific User by name
* `user_id` - (Optional) Returns information on a specific User by User id

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the User.
* `directory_user_id` - The identifier of the user account in the directory used for identity management.
* `hierarchy_group_id` - The identifier of the hierarchy group for the user.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the User separated by a colon (`:`).
* `identity_info` - A block that contains information about the identity of the user. [Documented below](#identity_info).
* `instance_id` - Specifies the identifier of the hosting Amazon Connect Instance.
* `phone_config` - A block that contains information about the phone settings for the user. [Documented below](#phone_config).
* `routing_profile_id` - The identifier of the routing profile for the user.
* `security_profile_ids` - A list of identifiers for the security profiles for the user.
* `tags` - A map of tags to assign to the User.

### `identity_info`

An `identity_info` block supports the following attributes:

* `email` - The email address.
* `first_name` - The first name.
* `last_name` - The last name.

### `phone_config`

A `phone_config` block supports the following attributes:

* `after_contact_work_time_limit` - The After Call Work (ACW) timeout setting, in seconds.
* `auto_accept` - When Auto-Accept Call is enabled for an available agent, the agent connects to contacts automatically.
* `desk_phone_number` - The phone number for the user's desk phone.
* `phone_type` - The phone type. Valid values are `DESK_PHONE` and `SOFT_PHONE`.
