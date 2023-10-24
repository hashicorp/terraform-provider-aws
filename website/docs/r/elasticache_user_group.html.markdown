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
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The user group identifier.
* `arn` - The ARN that identifies the user group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache user groups using the `user_group_id`. For example:

```terraform
import {
  to = aws_elasticache_user_group.my_user_group
  id = "userGoupId1"
}
```

Using `terraform import`, import ElastiCache user groups using the `user_group_id`. For example:

```console
% terraform import aws_elasticache_user_group.my_user_group userGoupId1
```
