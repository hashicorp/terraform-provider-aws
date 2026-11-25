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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN for the DB subnet group.
* `description` - Provides the description of the DB subnet group.
* `status` - Provides the status of the DB subnet group.
* `subnet_ids` - List of subnet identifiers.
* `supported_network_types` - Network type of the DB subnet group.
* `vpc_id` - Provides the VPC ID of the DB subnet group.
