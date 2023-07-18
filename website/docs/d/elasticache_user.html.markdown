---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user"
description: |-
  Get information on an ElastiCache User resource.
---

# Data Source: aws_elasticache_user

Use this data source to get information about an ElastiCache User.

## Example Usage

```hcl
data "aws_elasticache_user" "bar" {
  user_id = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `user_id` – (Required) Identifier for the user.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `user_id` - Identifier for the user.
* `user_name` - User name of the user.
* `access_string` - String for what access a user possesses within the associated ElastiCache replication groups or clusters.
