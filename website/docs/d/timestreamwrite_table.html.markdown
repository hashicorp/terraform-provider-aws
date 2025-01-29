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
  name          = aws_timestreamwrite_table.test.table_name
}
```

## Argument Reference

The following arguments are required:

* `database_name` - (Required) Name of the Timestream database.
* `name` - (Required) Name of the Timestream table.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN that uniquely identifies the table.
* `creation_time` - Time that table was created.
* `database_name` - Name of database.
* `last_updated_time` - Last time table was updated.
* `magnetic_store_write_properties` - Object containing the following attributes to desribe magnetic store writes.
    * `enable_magnetic_store_writes` - Flag that is set based on if magnetic store writes are enabled.
    * `magnetic_store_rejected_data_location` - Object containing the following attributes to describe error reports for records rejected during magnetic store writes.
        * `s3_configuration` - Object containing the following attributes to describe the configuration of an s3 location to write error reports for records rejected.
            * `bucket_name` - Name of S3 bucket.
            * `encryption_object` - Encryption option for  S3 location.
            * `kms_key_id` - AWS KMS key ID for S3 location with AWS maanged key.
            * `object_key_prefix` -  Object key preview for S3 location.
* `retention_properties` -  Object containing the following attributes to describe the retention duration for the memory and magnetic stores.
    * `magnetic_store_retention_period_in_days` - Duration in days in which the data must be stored in magnetic store.
    * `memory_store_retention_period_in_hours` - Duration in hours in which the data must be stored in memory store.
* `schema` -  Object containing the following attributes to describe the schema of the table.
    * `type` - Type of partition key.
    * `partition_key` - Level of enforcement for the specification of a dimension key in ingested records.
    * `name` - Name of the timestream attribute used for a dimension key.
* `name` - Name of the table.
* `table_status` - Current state of table.
