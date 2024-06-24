---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_replication_subnet_group"
description: |-
  Terraform data source for managing an AWS DMS (Database Migration) Replication Subnet Group.
---

# Data Source: aws_dms_replication_subnet_group

Terraform data source for managing an AWS DMS (Database Migration) Replication Subnet Group.

## Example Usage

### Basic Usage

```terraform
data "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
```

## Argument Reference

The following arguments are required:

* `replication_subnet_group_id` - (Required) Name for the replication subnet group. This value is stored as a lowercase string. It must contain no more than 255 alphanumeric characters, periods, spaces, underscores, or hyphens and cannot be `default`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `replication_subnet_group_description` - Description for the subnet group.
* `subnet_ids` - List of at least 2 EC2 subnet IDs for the subnet group. The subnets must cover at least 2 availability zones.
* `vpc_id` - The ID of the VPC the subnet group is in.
