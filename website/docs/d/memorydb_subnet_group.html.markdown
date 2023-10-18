---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_subnet_group"
description: |-
  Provides information about a MemoryDB Subnet Group.
---

# Resource: aws_memorydb_subnet_group

Provides information about a MemoryDB Subnet Group.

## Example Usage

```terraform
data "aws_memorydb_subnet_group" "example" {
  name = "my-subnet-group"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the subnet group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the subnet group.
* `arn` - ARN of the subnet group.
* `description` - Description of the subnet group.
* `subnet_ids` - Set of VPC Subnet ID-s of the subnet group.
* `vpc_id` - VPC in which the subnet group exists.
* `tags` - Map of tags assigned to the subnet group.
