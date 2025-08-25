---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_tables"
description: |-
  Get information on AWS Glue Data Catalog Tables within a database
---

# Data Source: aws_glue_catalog_tables

This data source can be used to fetch information about multiple AWS Glue Data Catalog Tables within a database.

## Example Usage

```terraform
data "aws_glue_catalog_tables" "example" {
  database_name = "my_database"
}
```

### Filter by table name expression

```terraform
data "aws_glue_catalog_tables" "example" {
  database_name = "my_database"
  expression    = "test_.*"
}
```

### Filter by table type

```terraform
data "aws_glue_catalog_tables" "example" {
  database_name = "my_database"
  table_type    = "EXTERNAL_TABLE"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `database_name` - (Required) Name of the metadata database where the table metadata resides.
* `catalog_id` - (Optional) ID of the Glue Catalog and database where the table metadata resides. If omitted, this defaults to the current AWS Account ID.
* `expression` - (Optional) A regular expression to filter the list of table names. Only table names matching this expression will be returned.
* `table_type` - (Optional) The type of tables to return. Valid values are `EXTERNAL_TABLE`, `MANAGED_TABLE`, `VIRTUAL_VIEW`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Catalog ID and Database name.
* `tables` - List of tables in the database. See [`tables`](#tables) below.

### tables

Each table in the list contains the following attributes:

* `arn` - The ARN of the Glue Table.
* `catalog_id` - ID of the Glue Catalog where the table resides.
* `database_name` - Name of the metadata database where the table metadata resides.
* `description` - Description of the table.
* `name` - Name of the table.
* `owner` - Owner of the table.
* `parameters` - Properties associated with this table, as a list of key-value pairs.
* `partition_keys` - Configuration block of columns by which the table is partitioned. Only primitive types are supported as partition keys. See [`partition_keys`](#partition_keys) below.
* `retention` - Retention time for this table.
* `storage_descriptor` - Configuration block for information about the physical storage of this table. For more information, refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor). See [`storage_descriptor`](#storage_descriptor) below.
* `table_type` - Type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
* `target_table` - Configuration block of a target table for resource linking. See [`target_table`](#target_table) below.
* `view_expanded_text` - If the table is a view, the expanded text of the view; otherwise null.
* `view_original_text` - If the table is a view, the original text of the view; otherwise null.

### partition_keys

* `comment` - Free-form text comment.
* `name` - Name of the Partition Key.
* `parameters` - Map of key-value pairs.
* `type` - Datatype of data in the Partition Key.

### storage_descriptor

* `additional_locations` - List of locations that point to the path where a Delta table is located
* `bucket_columns` - List of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `columns` - Configuration block for columns in the table. See [`columns`](#columns) below.
* `compressed` - Whether the data in the table is compressed.
* `input_format` - Input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `location` - Physical location of the table. By default, this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `number_of_buckets` - Is if the table contains any dimension columns.
* `output_format` - Output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `parameters` - User-supplied properties in key-value form.
* `schema_reference` - Object that references a schema stored in the AWS Glue Schema Registry. See [`schema_reference`](#schema_reference) below.
* `ser_de_info` - Configuration block for serialization and deserialization ("SerDe") information. See [`ser_de_info`](#ser_de_info) below.
* `skewed_info` - Configuration block with information about values that appear very frequently in a column (skewed values). See [`skewed_info`](#skewed_info) below.
* `sort_columns` - Configuration block for the sort order of each bucket in the table. See [`sort_columns`](#sort_columns) below.
* `stored_as_sub_directories` - Whether the table data is stored in subdirectories.

#### columns

* `comment` - Free-form text comment.
* `name` - Name of the Column.
* `parameters` - Key-value pairs defining properties associated with the column.
* `type` - Datatype of data in the Column.

#### schema_reference

* `schema_id` - Configuration block that contains schema identity fields. See [`schema_id`](#schema_id) below.
* `schema_version_id` - Unique ID assigned to a version of the schema.
* `schema_version_number` - Version number of the schema.

##### schema_id

* `registry_name` - Name of the schema registry that contains the schema.
* `schema_arn` - ARN of the schema.
* `schema_name` - Name of the schema.

#### ser_de_info

* `name` - Name of the SerDe.
* `parameters` - Map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - Usually the class that implements the SerDe. An example is `org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe`.

#### sort_columns

* `column` - Name of the column.
* `sort_order` - Whether the column is sorted in ascending (`1`) or descending order (`0`).

#### skewed_info

* `skewed_column_names` - List of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - List of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - Map of skewed values to the columns that contain them.

### target_table

* `catalog_id` - ID of the Data Catalog in which the table resides.
* `database_name` - Name of the catalog database that contains the target table.
* `name` - Name of the target table.
* `region` - Region of the target table.