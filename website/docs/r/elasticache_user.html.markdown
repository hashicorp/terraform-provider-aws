---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user"
description: |-
  Provides an ElastiCache user.
---

# Resource: aws_elasticache_user

Provides an ElastiCache user resource.

~> **Note:** All arguments including the username and passwords will be stored in the raw state as plain-text unless you use the write-only password arguments (`password_wo_1`/`password_wo_2` or `passwords_wo`).
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "testUserName"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "redis"
  passwords     = ["password123456789"]
}
```

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "testUserName"
  access_string = "on ~* +@all"
  engine        = "redis"

  authentication_mode {
    type = "iam"
  }
}
```

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "testUserName"
  access_string = "on ~* +@all"
  engine        = "redis"

  authentication_mode {
    type      = "password"
    passwords = ["password1", "password2"]
  }
}
```

### Using Write-Only Password (Terraform 1.11+)

~> **Note:** The top-level `passwords_wo` and `passwords_wo_version` attributes are deprecated. Use `password_wo_1`, `password_wo_2`, and `password_wo_version` instead for write-only password support with the ability to set up to 2 simultaneous passwords.

```terraform
resource "aws_elasticache_user" "test" {
  user_id              = "testUserId"
  user_name            = "testUserName"
  access_string        = "on ~* +@all"
  engine               = "redis"
  passwords_wo         = var.elasticache_password
  passwords_wo_version = 1 # Increment to trigger password update
}
```

### Using Write-Only Passwords in authentication_mode (Terraform 1.11+)

```terraform
resource "aws_elasticache_user" "test" {
  user_id             = "testUserId"
  user_name           = "testUserName"
  access_string       = "on ~* +@all"
  engine              = "redis"
  password_wo_1       = var.elasticache_password
  password_wo_version = 1 # Increment to trigger password update

  authentication_mode {
    type = "password"
  }
}
```

### Zero-Downtime Password Rotation (Terraform 1.11+)

ElastiCache supports up to 2 simultaneous valid passwords for zero-downtime rotation. Both passwords are equally valid — the numbering is for identification only, not priority.

```terraform
resource "aws_elasticache_user" "test" {
  user_id             = "testUserId"
  user_name           = "testUserName"
  access_string       = "on ~* +@all"
  engine              = "redis"
  password_wo_1       = var.current_password
  password_wo_2       = var.new_password
  password_wo_version = 2 # Increment to trigger password update

  authentication_mode {
    type = "password"
  }
}
```

## Argument Reference

The following arguments are required:

* `access_string` - (Required) Access permissions string used for this user. See [Specifying Permissions Using an Access String](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Clusters.RBAC.html#Access-string) for more details.
* `engine` - (Required) The current supported values are `redis`, `valkey` (case insensitive).
* `user_id` - (Required) The ID of the user.
* `user_name` - (Required) The username of the user.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `authentication_mode` - (Optional) Denotes the user's authentication properties. Detailed below.
* `no_password_required` - (Optional) Indicates a password is not required for this user.
* `passwords` - (Optional) Passwords used for this user. You can create up to two passwords for each user.
* `password_wo_1` - (Optional) First write-only password for this user. Not stored in state. Valid values are between 16 and 128 characters. Both `password_wo_1` and `password_wo_2` are equally valid — the numbering is for identification only, not priority. Conflicts with `passwords` and `passwords_wo`. Requires `password_wo_version`. Requires Terraform 1.11+.
* `password_wo_2` - (Optional) Second write-only password for this user. Not stored in state. Valid values are between 16 and 128 characters. Used alongside `password_wo_1` for zero-downtime password rotation. Conflicts with `passwords` and `passwords_wo`. Requires `password_wo_version`. Requires Terraform 1.11+.
* `password_wo_version` - (Optional) Version number for `password_wo_1`/`password_wo_2`. Increment this value to trigger a password update. Required when using `password_wo_1` or `password_wo_2`.
* `passwords_wo` - (Optional, **Deprecated**) Write-only password for this user. Use `password_wo_1` and `password_wo_2` instead. This argument is not stored in state. Conflicts with `passwords` and `authentication_mode`. See [Write-Only Arguments](https://developer.hashicorp.com/terraform/language/resources/syntax#write-only-arguments) for more information. Requires Terraform 1.11+.
* `passwords_wo_version` - (Optional, **Deprecated**) Version number for `passwords_wo`. Use `password_wo_version` instead. Increment this value to trigger a password update. Required when using `passwords_wo`.
* `tags` - (Optional) A list of tags to be added to this resource. A tag is a key-value pair.

### authentication_mode Configuration Block

* `passwords` - (Optional) Specifies the passwords to use for authentication if `type` is set to `password`. Conflicts with `password_wo_1` and `password_wo_2`.
* `type` - (Required) Specifies the authentication type. Possible options are: `password`, `no-password-required` or `iam`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the created ElastiCache User.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `read` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache users using the `user_id`. For example:

```terraform
import {
  to = aws_elasticache_user.my_user
  id = "userId1"
}
```

Using `terraform import`, import ElastiCache users using the `user_id`. For example:

```console
% terraform import aws_elasticache_user.my_user userId1
```
