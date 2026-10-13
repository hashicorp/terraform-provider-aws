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

    compaction_configuration {
      iceberg_configuration {
        strategy              = "binpack"
        min_input_files       = 5
        delete_file_threshold = 1
      }
    }
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

* `catalog_id` - (Required) Catalog ID of the table.
* `configuration` - (Required) Configuration block that defines the table optimizer settings. See [`configuration` Block](#configuration-block) for additional details.
* `database_name` - (Required) Name of the database in the catalog in which the table resides.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_name` - (Required) Name of the table.
* `type` - (Required) Type of table optimizer. Valid values are `compaction`, `retention`, and `orphan_file_deletion`.

### `configuration` Block

* `compaction_configuration` - (Optional) Configuration block for a compaction optimizer. See [`compaction_configuration` Block](#compaction_configuration-block) for additional details.
* `enabled` - (Required) Whether the table optimizer is enabled.
* `orphan_file_deletion_configuration` - (Optional) Configuration block for an orphan file deletion optimizer. See [`orphan_file_deletion_configuration` Block](#orphan_file_deletion_configuration-block) for additional details.
* `retention_configuration` - (Optional) Configuration block for a snapshot retention optimizer. See [`retention_configuration` Block](#retention_configuration-block) for additional details.
* `role_arn` - (Required) ARN of the IAM role to use for the table optimizer.

#### `compaction_configuration` Block

* `iceberg_configuration` - (Optional) Configuration for an Iceberg compaction optimizer. See [`configuration.compaction_configuration.iceberg_configuration` Block](#configurationcompaction_configurationiceberg_configuration-block) for additional details.

##### `configuration.compaction_configuration.iceberg_configuration` Block

* `delete_file_threshold` - (Optional) Minimum number of deletes that must be present in a data file to make it eligible for compaction. Defaults to `1`.
* `min_input_files` - (Optional) Minimum number of data files that must be present in a partition before compaction will actually compact files. Defaults to `100`.
* `strategy` - (Optional) Strategy to use for compaction. Valid values are `binpack`, `sort` and `z-order` Defaults to `binpack`.

#### `orphan_file_deletion_configuration` Block

* `iceberg_configuration` - (Optional) Configuration for an Iceberg orphan file deletion optimizer. See [`configuration.orphan_file_deletion_configuration.iceberg_configuration` Block](#configurationorphan_file_deletion_configurationiceberg_configuration-block) for additional details.

##### `configuration.orphan_file_deletion_configuration.iceberg_configuration` Block

* `location` - (Optional) Directory in which to look for files. You may choose a sub-directory rather than the top-level table location. Defaults to the table's location.
* `orphan_file_retention_period_in_days` - (Optional) Number of days that orphan files should be retained before file deletion. Defaults to `3`.
* `run_rate_in_hours` - (Optional) Interval in hours between orphan file deletion job runs. Defaults to `24`.

#### `retention_configuration` Block

* `iceberg_configuration` - (Optional) Configuration for an Iceberg snapshot retention optimizer. See [`configuration.retention_configuration.iceberg_configuration` Block](#configurationretention_configurationiceberg_configuration-block) for additional details.

##### `configuration.retention_configuration.iceberg_configuration` Block

* `clean_expired_files` - (Optional) If set to `false`, snapshots are only deleted from table metadata, and the underlying data and metadata files are not deleted. Defaults to `false`.
* `number_of_snapshots_to_retain` - (Optional) Number of Iceberg snapshots to retain within the retention period. Defaults to `1` or the corresponding Iceberg table configuration field if it exists.
* `run_rate_in_hours` - (Optional) Interval in hours between retention job runs. Defaults to `24`.
* `snapshot_retention_period_in_days` - (Optional) Number of days to retain the Iceberg snapshots. Defaults to `5`, or the corresponding Iceberg table configuration field if it exists.

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
