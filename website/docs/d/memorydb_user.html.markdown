---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_user"
description: |-
  Provides information about a MemoryDB User.
---

# Resource: aws_memorydb_user

Provides information about a MemoryDB User.

## Example Usage

```terraform
data "aws_memorydb_user" "example" {
  user_name = "my-user"
}
```

## Argument Reference

The following arguments are required:

* `user_name` - (Required) Name of the user.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the user.
* `access_string` - Access permissions string used for this user.
* `arn` - ARN of the user.
* `authentication_mode` - Denotes the user's authentication properties.
    * `password_count` - Number of passwords belonging to the user if `type` is set to `password`.
    * `type` - Type of authentication configured.
* `minimum_engine_version` - Minimum engine version supported for the user.
* `tags` - Map of tags assigned to the user.
