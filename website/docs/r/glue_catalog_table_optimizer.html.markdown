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

### Compaction Optimizer

```terraform
resource "aws_glue_catalog_table_optimizer" "example" {
  catalog_id    = "123456789012"
  database_name = "example_database"
  table_name    = "example_table"

  configuration {
    role_arn = "arn:aws:iam::123456789012:role/example-role"
    enabled  = true
  }

  type = "compaction"
}
```

### Snapshot Retention Optimizer

```terraform
resource "aws_glue_catalog_table_optimizer" "example" {
  catalog_id    = "123456789012"
  database_name = "example_database"
  table_name    = "example_table"

  configuration {
    role_arn = "arn:aws:iam::123456789012:role/example-role"
    enabled  = true

    retention_configuration {
      iceberg_configuration {
        snapshot_retention_period_in_days = 7
        number_of_snapshots_to_retain     = 3
        clean_expired_files               = true
      }
    }

  }

  type = "retention"
}
```

### Orphan File Deletion Optimizer

```terraform
resource "aws_glue_catalog_table_optimizer" "example" {
  catalog_id    = "123456789012"
  database_name = "example_database"
  table_name    = "example_table"

  configuration {
    role_arn = "arn:aws:iam::123456789012:role/example-role"
    enabled  = true

    orphan_file_deletion_configuration {
      iceberg_configuration {
        orphan_file_retention_period_in_days = 7
        location                             = "s3://example-bucket/example_table/"
      }
    }

  }

  type = "orphan_file_deletion"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `catalog_id` - (Required) The Catalog ID of the table.
* `configuration` - (Required) A configuration block that defines the table optimizer settings. See [Configuration](#configuration) for additional details.
* `database_name` - (Required) The name of the database in the catalog in which the table resides.
* `table_name` - (Required) The name of the table.
* `type` - (Required) The type of table optimizer. Valid values are `compaction`, `retention`, and `orphan_file_deletion`.

### Configuration

* `enabled` - (Required) Indicates whether the table optimizer is enabled.
* `orphan_file_deletion_configuration` (Optional) - The configuration block for an orphan file deletion optimizer. See [Orphan File Deletion Configuration](#orphan-file-deletion-configuration) for additional details.
* `retention_configuration` (Optional) - The configuration block for a snapshot retention optimizer. See [Retention Configuration](#retention-configuration) for additional details.
* `role_arn` - (Required) The ARN of the IAM role to use for the table optimizer.

### Orphan File Deletion Configuration

* `iceberg_configuration` (Optional) - The configuration for an Iceberg orphan file deletion optimizer.
    * `orphan_file_retention_period_in_days` (Optional) - The number of days that orphan files should be retained before file deletion. Defaults to `3`.
    * `location` (Optional) - Specifies a directory in which to look for files. You may choose a sub-directory rather than the top-level table location. Defaults to the table's location.
  
### Retention Configuration

* `iceberg_configuration` (Optional) - The configuration for an Iceberg snapshot retention optimizer.
    * `snapshot_retention_period_in_days` (Optional) - The number of days to retain the Iceberg snapshots. Defaults to `5`, or the corresponding Iceberg table configuration field if it exists.
    * `number_of_snapshots_to_retain` (Optional) - The number of Iceberg snapshots to retain within the retention period. Defaults to `1` or the corresponding Iceberg table configuration field if it exists.
    * `clean_expired_files` (Optional) - If set to `false`, snapshots are only deleted from table metadata, and the underlying data and metadata files are not deleted. Defaults to `false`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Catalog Table Optimizer using the `catalog_id,database_name,table_name,type`. For example:

```terraform
import {
  to = aws_glue_catalog_table_optimizer.example
  id = "123456789012,example_database,example_table,compaction"
}
```

Using `terraform import`, import Glue Catalog Table Optimizer using the `catalog_id,database_name,table_name,type`. For example:

```console
% terraform import aws_glue_catalog_table_optimizer.example 123456789012,example_database,example_table,compaction
```
