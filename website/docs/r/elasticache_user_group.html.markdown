---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user_group"
description: |-
  Provides an ElastiCache user group.
---

# Resource: aws_elasticache_user_group

Provides an ElastiCache user group resource.

## Example Usage

```terraform
resource "aws_elasticache_user" "test" {
  user_id       = "testUserId"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  engine        = "REDIS"
  user_group_id = "userGroupId"
  user_ids      = [aws_elasticache_user.test.user_id]
}
```

## Argument Reference

The following arguments are required:

* `engine` - (Required) The current supported value is `REDIS`.
* `user_group_id` - (Required) The ID of the user group.

The following arguments are optional:

* `user_ids` - (Optional) The list of user IDs that belong to the user group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The user group identifier.

## Import

ElastiCache user groups can be imported using the `user_group_id`, e.g.,

```
$ terraform import aws_elasticache_user_group.my_user_group userGoupId1
```
