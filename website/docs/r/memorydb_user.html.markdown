---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_user"
description: |-
  Provides a MemoryDB User.
---

# Resource: aws_memorydb_user

Provides a MemoryDB User.

More information about users and ACL-s can be found in the [MemoryDB User Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/clusters.acls.html).

~> **Note:** All arguments including the username and passwords will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "random_password" "example" {
  length = 16
}

resource "aws_memorydb_user" "example" {
  user_name     = "my-user"
  access_string = "on ~* &* +@all"

  authentication_mode {
    type      = "password"
    passwords = [random_password.example.result]
  }
}
```

## Argument Reference

The following arguments are required:

* `access_string` - (Required) Access permissions string used for this user.
* `authentication_mode` - (Required) Denotes the user's authentication properties. Detailed below.
* `user_name` - (Required, Forces new resource) Name of the MemoryDB user. Up to 40 characters.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### authentication_mode Configuration Block

* `passwords` - (Optional) Set of passwords used for authentication if `type` is set to `password`. You can create up to two passwords for each user.
* `type` - (Required) Specifies the authentication type. Valid values are: `password` or `iam`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Same as `user_name`.
* `arn` - ARN of the user.
* `minimum_engine_version` - Minimum engine version supported for the user.
* `authentication_mode` configuration block
    * `password_count` - Number of passwords belonging to the user if `type` is set to `password`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a user using the `user_name`. For example:

```terraform
import {
  to = aws_memorydb_user.example
  id = "my-user"
}
```

Using `terraform import`, import a user using the `user_name`. For example:

```console
% terraform import aws_memorydb_user.example my-user
```

The `passwords` are not available for imported resources, as this information cannot be read back from the MemoryDB API.
