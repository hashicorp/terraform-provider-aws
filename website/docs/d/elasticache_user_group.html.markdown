---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user_group"
description: |-
  Provides details about an AWS ElastiCache User Group.
---

# Data Source: aws_elasticache_user_group

Provides details about an AWS ElastiCache User Group.

## Example Usage

### Basic Usage

```terraform
data "aws_elasticache_user_group" "example" {
  user_group_id = "userGroupId"
}
```

## Argument Reference

The following arguments are required:

* `user_group_id` - (Required) ID of the user group.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the user group.
* `engine` - Cache engine for the user group.
* `id` - ID of the user group.
* `tags` - Map of tags assigned to the resource.
* `user_ids` - List of user IDs that belong to the user group.
