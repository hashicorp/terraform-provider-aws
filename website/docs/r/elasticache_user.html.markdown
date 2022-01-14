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

## Argument Reference

The following arguments are required:

* `access_string` - (Required) Access permissions string used for this user. See [Specifying Permissions Using an Access String](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Clusters.RBAC.html#Access-string) for more details.
* `engine` - (Required) The current supported value is `REDIS`.
* `user_id` - (Required) The ID of the user.
* `user_name` - (Required) The username of the user.

The following arguments are optional:

* `no_password_required` - (Optional) Indicates a password is not required for this user.
* `passwords` - (Optional) Passwords used for this user. You can create up to two passwords for each user.
* `tags` - (Optional) A list of tags to be added to this resource. A tag is a key-value pair.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the created ElastiCache User.

## Import

ElastiCache users can be imported using the `user_id`, e.g.,

```
$ terraform import aws_elasticache_user.my_user userId1
```
