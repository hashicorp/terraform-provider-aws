---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_subnet_group"
description: |-
  Provides information about a ElastiCache Subnet Group.
---

# Resource: aws_elasticache_subnet_group

Provides information about a ElastiCache Subnet Group.

## Example Usage

```terraform
data "aws_elasticache_subnet_group" "example" {
  name = "my-subnet-group"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the subnet group.

## Attributes Reference

In addition, the following attributes are exported:

* `id` - Name of the subnet group.
* `arn` - ARN of the subnet group.
* `description` - Description of the subnet group.
* `subnet_ids` - Set of VPC Subnet ID-s of the subnet group.
* `tags` - Map of tags assigned to the subnet group.
