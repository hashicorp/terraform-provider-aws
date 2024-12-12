---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user"
description: |-
  Provides an ElastiCache user.
---

# Resource: aws_elasticache_user

Provides an ElastiCache user resource.

~> **Note:** All arguments including the username and passwords will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "testUserName"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
```

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "testUserName"
  access_string = "on ~* +@all"
  engine        = "REDIS"

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
  engine        = "REDIS"

  authentication_mode {
    type      = "password"
    passwords = ["password1", "password2"]
  }
}
```

## Argument Reference

The following arguments are required:

* `access_string` - (Required) Access permissions string used for this user. See [Specifying Permissions Using an Access String](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Clusters.RBAC.html#Access-string) for more details.
* `engine` - (Required) The current supported value is `REDIS`.
* `user_id` - (Required) The ID of the user.
* `user_name` - (Required) The username of the user.

The following arguments are optional:

* `authentication_mode` - (Optional) Denotes the user's authentication properties. Detailed below.
* `no_password_required` - (Optional) Indicates a password is not required for this user.
* `passwords` - (Optional) Passwords used for this user. You can create up to two passwords for each user.
* `tags` - (Optional) A list of tags to be added to this resource. A tag is a key-value pair.

### authentication_mode Configuration Block

* `passwords` - (Optional) Specifies the passwords to use for authentication if `type` is set to `password`.
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
