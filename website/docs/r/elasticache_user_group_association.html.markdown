---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user_group_association"
description: |-
  Associate an ElastiCache user and user group.
---

# Resource: aws_elasticache_user_group_association

Associate an existing ElastiCache user and an existing user group.

~> **NOTE:** Terraform will detect changes in the `aws_elasticache_user_group` since `aws_elasticache_user_group_association` changes the user IDs associated with the user group. You can ignore these changes with the `lifecycle` `ignore_changes` meta argument as shown in the example.

## Example Usage

```terraform
resource "aws_elasticache_user" "default" {
  user_id       = "defaultUserID"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "example" {
  engine        = "REDIS"
  user_group_id = "userGroupId"
  user_ids      = [aws_elasticache_user.default.user_id]

  lifecycle {
    ignore_changes = [user_ids]
  }
}

resource "aws_elasticache_user" "example" {
  user_id       = "exampleUserID"
  user_name     = "exampleuser"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group_association" "example" {
  user_group_id = aws_elasticache_user_group.example.user_group_id
  user_id       = aws_elasticache_user.example.user_id
}
```

## Argument Reference

The following arguments are required:

* `user_group_id` - (Required) ID of the user group.
* `user_id` - (Required) ID of the user to associated with the user group.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache user group associations using the `user_group_id` and `user_id`. For example:

```terraform
import {
  to = aws_elasticache_user_group_association.example
  id = "userGoupId1,userId"
}
```

Using `terraform import`, import ElastiCache user group associations using the `user_group_id` and `user_id`. For example:

```console
% terraform import aws_elasticache_user_group_association.example userGoupId1,userId
```
