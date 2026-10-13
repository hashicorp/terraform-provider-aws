---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_partition"
description: |-
  Provides a Glue Partition.
---

# Resource: aws_glue_partition

Provides a Glue Partition Resource.

## Example Usage

```terraform
resource "aws_glue_partition" "example" {
  database_name    = "some-database"
  table_name       = "some-table"
  partition_values = ["some-value"]
}
```

## Argument Reference

This resource supports the following arguments:

* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.
* `parameters` - (Optional) Properties associated with this table, as a list of key-value pairs.
* `partition_values` - (Required) Values that define the partition.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_descriptor` - (Optional) [Storage descriptor](#storage_descriptor-block) object containing information about the physical storage of this table. You can refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor) for a full explanation of this object.
* `table_name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.

### `storage_descriptor` Block

* `additional_locations` - (Optional) List of locations that point to the path where a Delta table is located.
* `bucket_columns` - (Optional) List of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `columns` - (Optional) List of the [`columns`](#columns-block) in the table.
* `compressed` - (Optional) True if the data in the table is compressed, or False if not.
* `input_format` - (Optional) Input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `location` - (Optional) Physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `output_format` - (Optional) Output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `ser_de_info` - (Optional) [Serialization/deserialization (SerDe)](#ser_de_info-block) information.
* `skewed_info` - (Optional) Information about values that appear very frequently in a column (skewed values).
* `sort_columns` - (Optional) List of [`sort_columns`](#sort_columns-block) objects specifying the sort order of each bucket in the table.
* `stored_as_sub_directories` - (Optional) True if the table data is stored in subdirectories, or False if not.

### `columns` Block

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Column.
* `type` - (Optional) Datatype of data in the Column.

### `ser_de_info` Block

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) Map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is: org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe.

### `sort_columns` Block

* `column` - (Required) Name of the column.
* `sort_order` - (Required) Order the column is sorted in: ascending (`1`) or descending (`0`).

### `skewed_info` Block

* `skewed_column_names` - (Optional) List of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) List of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) Map of skewed values to the columns that contain them.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - Time at which the partition was created.
* `id` - partition ID.
* `last_accessed_time` - Last time at which the partition was accessed.
* `last_analyzed_time` - Last time at which column statistics were computed for this partition.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Partitions using the catalog ID (usually AWS account ID), database name, table name and partition values. For example:

```terraform
import {
  to = aws_glue_partition.part
  id = "123456789012:MyDatabase:MyTable:val1#val2"
}
```

Using `terraform import`, import Glue Partitions using the catalog ID (usually AWS account ID), database name, table name and partition values. For example:

```console
% terraform import aws_glue_partition.part 123456789012:MyDatabase:MyTable:val1#val2
```
