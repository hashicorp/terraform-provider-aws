---
subcategory: "Timestream Write"
layout: "aws"
page_title: "AWS: aws_timestreamwrite_table"
description: |-
  Provides a Timestream table resource.
---

# Resource: aws_timestreamwrite_table

Provides a Timestream table resource.

## Example Usage

### Basic usage

```terraform
resource "aws_timestreamwrite_table" "example" {
  database_name = aws_timestreamwrite_database.example.database_name
  table_name    = "example"
}
```

### Full usage

```terraform
resource "aws_timestreamwrite_table" "example" {
  database_name = aws_timestreamwrite_database.example.database_name
  table_name    = "example"

  retention_properties {
    magnetic_store_retention_period_in_days = 30
    memory_store_retention_period_in_hours  = 8
  }

  tags = {
    Name = "example-timestream-table"
  }
}
```

### Customer-defined Partition Key

```terraform
resource "aws_timestreamwrite_table" "example" {
  database_name = aws_timestreamwrite_database.example.database_name
  table_name    = "example"

  schema {
    composite_partition_key {
      enforcement_in_record = "REQUIRED"
      name                  = "attr1"
      type                  = "DIMENSION"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `database_name` - (Required) Name of the Timestream database.
* `magnetic_store_write_properties` - (Optional) Properties to set on the table when enabling magnetic store writes. See [`magnetic_store_write_properties` Block](#magnetic_store_write_properties-block) below for more details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention_properties` - (Optional) Retention duration for the memory store and magnetic store. See [`retention_properties` Block](#retention_properties-block) below for more details. If not provided, `magnetic_store_retention_period_in_days` default to 73000 and `memory_store_retention_period_in_hours` defaults to 6.
* `schema` - (Optional) Schema of the table. See [`schema` Block](#schema-block) below for more details.
* `table_name` - (Required) Name of the Timestream table.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `magnetic_store_write_properties` Block

The `magnetic_store_write_properties` block supports the following arguments:

* `enable_magnetic_store_writes` - (Optional) Flag to enable magnetic store writes.
* `magnetic_store_rejected_data_location` - (Optional) Location to write error reports for records rejected asynchronously during magnetic store writes. See [`magnetic_store_rejected_data_location` Block](#magnetic_store_rejected_data_location-block) below for more details.

#### `magnetic_store_rejected_data_location` Block

The `magnetic_store_rejected_data_location` block supports the following arguments:

* `s3_configuration` - (Optional) Configuration of an S3 location to write error reports for records rejected, asynchronously, during magnetic store writes. See [`s3_configuration` Block](#s3_configuration-block) below for more details.

#### `s3_configuration` Block

The `s3_configuration` block supports the following arguments:

* `bucket_name` - (Optional) Bucket name of the customer S3 bucket.
* `encryption_option` - (Optional) Encryption option for the customer s3 location. Options are S3 server side encryption with an S3-managed key or KMS managed key. Valid values are `SSE_KMS` and `SSE_S3`.
* `kms_key_id` - (Optional) KMS key arn for the customer s3 location when encrypting with a KMS managed key.
* `object_key_prefix` - (Optional) Object key prefix for the customer S3 location.

### `retention_properties` Block

The `retention_properties` block supports the following arguments:

* `magnetic_store_retention_period_in_days` - (Required) Duration for which data must be stored in the magnetic store. Minimum value of 1. Maximum value of 73000.
* `memory_store_retention_period_in_hours` - (Required) Duration for which data must be stored in the memory store. Minimum value of 1. Maximum value of 8766.

### `schema` Block

The `schema` block supports the following arguments:

* `composite_partition_key` - (Required) Non-empty list of partition keys defining the attributes used to partition the table data. The order of the list determines the partition hierarchy. The name and type of each partition key as well as the partition key order cannot be changed after the table is created. However, the enforcement level of each partition key can be changed. See [`composite_partition_key` Block](#composite_partition_key-block) below for more details.

### `composite_partition_key` Block

The `composite_partition_key` block supports the following arguments:

* `enforcement_in_record` - (Optional) Level of enforcement for the specification of a dimension key in ingested records. Valid values: `REQUIRED`, `OPTIONAL`.
* `name` - (Optional) Name of the attribute used for a dimension key.
* `type` - (Required) Type of the partition key. Valid values: `DIMENSION`, `MEASURE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN that uniquely identifies this table.
* `id` - `table_name` and `database_name` separated by a colon (`:`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Timestream tables using the `table_name` and `database_name` separate by a colon (`:`). For example:

```terraform
import {
  to = aws_timestreamwrite_table.example
  id = "ExampleTable:ExampleDatabase"
}
```

Using `terraform import`, import Timestream tables using the `table_name` and `database_name` separate by a colon (`:`). For example:

```console
% terraform import aws_timestreamwrite_table.example ExampleTable:ExampleDatabase
```
