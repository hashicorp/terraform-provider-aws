---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_subnet_group"
description: |-
  Get information on an RDS Database Subnet Group.
---

# Data Source: aws_db_subnet_group

Use this data source to get information about an RDS subnet group.

## Example Usage

```terraform
data "aws_db_subnet_group" "database" {
  name = "my-test-database-subnet-group"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the RDS database subnet group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN for the DB subnet group.
* `description` - Provides the description of the DB subnet group.
* `status` - Provides the status of the DB subnet group.
* `subnet_ids` - Contains a list of subnet identifiers.
* `supported_network_types` - The network type of the DB subnet group.
* `vpc_id` - Provides the VPC ID of the DB subnet group.
