---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_table_optimizer"
description: |-
  Terraform resource for managing an AWS Glue Catalog Table Optimizer.
---

# Resource: aws_glue_catalog_table_optimizer

Terraform resource for managing an AWS Glue Catalog Table Optimizer.

## Example Usage

### Basic Usage

```terraform
resource "aws_glue_catalog_table_optimizer" "example" {
  catalog_id     = "123456789012"
  database_name  = "test_database"
  table_name     = "test_table"
  configuration  = {
    role_arn = "arn:aws:iam::123456789012:role/example-role"
    enabled  = true
  }
  type = "compaction"
}

## Argument Reference

The following arguments are required:

* `catalog_id` - (Required) The Catalog ID of the table.

* `database_name` - (Required) The name of the database in the catalog in which the table resides.

* `table_name` - (Required) The name of the table.

* `type` - (Required) The type of table optimizer. Currently, the only valid value is compaction.

* `configuration` - (Required) A configuration block that defines the table optimizer settings. The block contains:
  
    * `role_arn` - (Required) The ARN of the IAM role to use for the table optimizer.
  
    * `enabled` - (Required) Indicates whether the table optimizer is enabled.
