---
subcategory: "Timestream Write"
layout: "aws"
page_title: "AWS: aws_timestreamwrite_table"
description: |-
  Terraform data source for managing an AWS Timestream Write Table.
---

# Data Source: aws_timestreamwrite_table

Terraform data source for managing an AWS Timestream Write Table.

## Example Usage

### Basic Usage

```terraform
data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
```

## Argument Reference

The following arguments are required:

* `database_name` - (Required) The name of the Timestream database.

* `table_name` - (Required) The name of the Timestream table.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The `table_name` and `database_name` separated by a colon (`:`).
* `arn` - The ARN that uniquely identifies this table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).