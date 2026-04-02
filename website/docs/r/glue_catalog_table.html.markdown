---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_table"
description: |-
  Provides a Glue Catalog Table.
---

# Resource: aws_glue_catalog_table

Provides a Glue Catalog Table Resource. You can refer to the [Glue Developer Guide](http://docs.aws.amazon.com/glue/latest/dg/populate-data-catalog.html) for a full explanation of the Glue Data Catalog functionality.

## Example Usage

### Basic Table

```terraform
resource "aws_glue_catalog_table" "example" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"
}
```

### Parquet Table for Athena

```terraform
resource "aws_glue_catalog_table" "example" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"

  table_type = "EXTERNAL_TABLE"

  parameters = {
    EXTERNAL              = "TRUE"
    "parquet.compression" = "SNAPPY"
  }

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"

      parameters = {
        "serialization.format" = 1
      }
    }

    columns {
      name = "my_string"
      type = "string"
    }

    columns {
      name = "my_double"
      type = "double"
    }

    columns {
      name    = "my_date"
      type    = "date"
      comment = ""
    }

    columns {
      name    = "my_bigint"
      type    = "bigint"
      comment = ""
    }

    columns {
      name    = "my_struct"
      type    = "struct<my_nested_string:string>"
      comment = ""
    }
  }
}
```

### Iceberg Table

```terraform
resource "aws_glue_catalog_table" "example" {
  name          = "transactiontable1"
  database_name = "bankdata_icebergdb"

  open_table_format_input {
    iceberg_input {
      metadata_operation = "CREATE"
      version            = 2

      iceberg_table_input {
        location = "s3://sampledatabucket/bankdataiceberg/transactiontable1/"

        schema {
          schema_id = 0
          type      = "struct"

          fields {
            id       = 1
            name     = "transaction_id"
            required = true
            type     = <<EOF
            "string"
EOF
          }
          fields {
            id       = 2
            name     = "transaction_date"
            required = true
            type     = <<EOF
            "date"
EOF
          }
          fields {
            id       = 3
            name     = "monthly_balance"
            required = true
            type     = <<EOF
            "float"
EOF
          }
        }

        partition_spec {
          fields {
            name      = "by_year"
            source_id = 2
            transform = "year"
          }

          spec_id = 0
        }

        sort_order {
          fields {
            direction  = "asc"
            null_order = "nulls-last"
            source_id  = 1
            transform  = "none"
          }

          order_id = 1
        }
      }
    }
  }
}
```

### Protected View

```terraform
resource "aws_glue_catalog_table" "example" {
  name          = "multidialect_view"
  database_name = "catalog_database"

  table_type = "VIRTUAL_VIEW"

  view_definition {
    is_protected = true

    representations {
      dialect               = "ATHENA"
      dialect_version       = "3"
      view_original_text    = "SELECT * FROM catalog_database.base_table"
      validation_connection = aws_glue_connection.example.name
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.
* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `description` - (Optional) Description of the table.
* `owner` - (Optional) Owner of the table.
* `open_table_format_input` - (Optional) Configuration block for open table formats. See [`open_table_format_input`](#open_table_format_input) below.
* `parameters` - (Optional) Properties associated with this table, as a list of key-value pairs.
* `partition_index` - (Optional) Configuration block for a maximum of 3 partition indexes. See [`partition_index`](#partition_index) below.
* `partition_keys` - (Optional) Configuration block of columns by which the table is partitioned. Only primitive types are supported as partition keys. See [`partition_keys`](#partition_keys) below.
* `retention` - (Optional) Retention time for this table.
* `storage_descriptor` - (Optional) Configuration block for information about the physical storage of this table. For more information, refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor). See [`storage_descriptor`](#storage_descriptor) below.
* `table_type` - (Optional) Type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
* `target_table` - (Optional) Configuration block of a target table for resource linking. See [`target_table`](#target_table) below.
* `view_expanded_text` - (Optional) If the table is a view, the expanded text of the view; otherwise null.
* `view_original_text` - (Optional) If the table is a view, the original text of the view; otherwise null.
* `view_definition` - (Optional) A structure that contains all the information that defines the view, including the dialect or dialects for the view, and the query. See [`view_definition`](#view_definition) below.

### open_table_format_input

~> **NOTE:** A `open_table_format_input` cannot be added to an existing `glue_catalog_table`.
This will destroy and recreate the table, possibly resulting in data loss.

* `iceberg_input` - (Required) Configuration block for iceberg table config. See [`iceberg_input`](#iceberg_input) below.

### iceberg_input

~> **NOTE:** A `iceberg_input` cannot be added to an existing `open_table_format_input`.
This will destroy and recreate the table, possibly resulting in data loss.

* `iceberg_table_input` - (Optional) Configuration parameters, including table properties and metadata specifications. See [`iceberg_table_input`](#iceberg_table_input) below.
* `metadata_operation` - (Required) A required metadata operation. Can only be set to CREATE.
* `version` - (Optional) The table version for the Iceberg table. Defaults to 2.

### iceberg_table_input

* `location` - (Required) The S3 location where the Iceberg table data will be stored. Maximum length of 2056 characters.
* `partition_spec` - (Optional) The partitioning specification that defines how the Iceberg table data will be organized and partitioned for optimal query performance. See [`partition_spec`](#partition_spec) below.
* `properties` - (Optional) Key-value pairs of additional table properties and configuration settings for the Iceberg table.
* `schema` - (Required) The schema definition that specifies the structure, field types, and metadata for the Iceberg table. See [`schema`](#schema) below.
* `sort_order` - (Optional) The sort order specification that defines how data should be ordered within each partition to optimize query performance. See [`sort_order`](#sort_order) below.

### partition_spec

* `fields` - (Required) The list of partition fields that define how the table data should be partitioned. See [`fields`](#partition-fields) below.
* `spec_id` - (Optional) The unique identifier for this partition specification within the Iceberg table's metadata history.

#### partition fields

* `field_id` - (Optional) The unique identifier assigned to this partition field within the Iceberg table's partition specification.
* `name` - (Required) The name of the partition field as it will appear in the partitioned table structure. Length between 1 and 1024 characters.
* `source_id` - (Required) The identifier of the source field from the table schema that this partition field is based on.
* `transform` - (Required) The transformation function applied to the source field to create the partition. Common values: `identity`, `bucket`, `truncate`, `year`, `month`, `day`, `hour`.

### schema

* `fields` - (Required) The list of field definitions that make up the table schema. See [`fields`](#schema-fields) below.
* `identifier_field_ids` - (Optional) The list of field identifiers that uniquely identify records in the table, used for row-level operations and deduplication.
* `schema_id` - (Optional) The unique identifier for this schema version within the Iceberg table's schema evolution history.
* `type` - (Optional) The root type of the schema structure. Valid value: `struct`.

#### schema fields

* `doc` - (Optional) Optional documentation or description text that provides additional context about the purpose and usage of this field. Length between 0 and 255 characters.
* `id` - (Required) The unique identifier assigned to this field within the Iceberg table schema, used for schema evolution and field tracking.
* `initial_default` - (Optional) Default value as JSON used to populate the field's value for all records that were written before the field was added to the schema.
* `name` - (Required) The name of the field as it appears in the table schema and query operations. Length between 1 and 1024 characters.
* `required` - (Required) Indicates whether this field is required (non-nullable) or optional (nullable) in the table schema.
* `type` - (Required) The data type definition for this field as a JSON string, specifying the structure and format of the data it contains. Examples: `"long"`, `"string"`, `"timestamp"`, `"decimal(10,2)"`.
* `write_default` - (Optional) Default value as JSON used to populate the field's value for any records written after the field was added to the schema, if the writer does not supply the field's value.

### sort_order

* `fields` - (Required) The list of fields and their sort directions that define the ordering criteria for the Iceberg table data. See [`fields`](#sort-fields) below.
* `order_id` - (Required) The unique identifier for this sort order specification within the Iceberg table's metadata.

#### sort fields

* `direction` - (Required) The sort direction for this field. Valid values: `asc`, `desc`.
* `null_order` - (Required) The ordering behavior for null values in this field. Valid values: `nulls-first`, `nulls-last`.
* `source_id` - (Required) The identifier of the source field from the table schema that this sort field is based on.
* `transform` - (Required) The transformation function applied to the source field before sorting. Common values: `identity`, `bucket`, `truncate`.

### partition_index

~> **NOTE:** A `partition_index` cannot be added to an existing `glue_catalog_table`.
This will destroy and recreate the table, possibly resulting in data loss.
To add an index to an existing table, see the [`glue_partition_index` resource](/docs/providers/aws/r/glue_partition_index.html) for configuration details.

* `index_name` - (Required) Name of the partition index.
* `keys` - (Required) Keys for the partition index.

### partition_keys

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Partition Key.
* `parameters` - (Optional) Map of key-value pairs.
* `type` - (Optional) Datatype of data in the Partition Key.

### storage_descriptor

* `additional_locations` - (Optional) List of locations that point to the path where a Delta table is located.
* `bucket_columns` - (Optional) List of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `columns` - (Optional) Configuration block for columns in the table. See [`columns`](#columns) below.
* `compressed` - (Optional) Whether the data in the table is compressed.
* `input_format` - (Optional) Input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `location` - (Optional) Physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `output_format` - (Optional) Output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `schema_reference` - (Optional) Object that references a schema stored in the AWS Glue Schema Registry. When creating a table, you can pass an empty list of columns for the schema, and instead use a schema reference. See [Schema Reference](#schema_reference) below.
* `ser_de_info` - (Optional) Configuration block for serialization and deserialization ("SerDe") information. See [`ser_de_info`](#ser_de_info) below.
* `skewed_info` - (Optional) Configuration block with information about values that appear very frequently in a column (skewed values). See [`skewed_info`](#skewed_info) below.
* `sort_columns` - (Optional) Configuration block for the sort order of each bucket in the table. See [`sort_columns`](#sort_columns) below.
* `stored_as_sub_directories` - (Optional) Whether the table data is stored in subdirectories.

#### columns

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Column.
* `parameters` - (Optional) Key-value pairs defining properties associated with the column.
* `type` - (Optional) Datatype of data in the Column.

#### schema_reference

* `schema_id` - (Optional) Configuration block that contains schema identity fields. Either this or the `schema_version_id` has to be provided. See [`schema_id`](#schema_id) below.
* `schema_version_id` - (Optional) Unique ID assigned to a version of the schema. Either this or the `schema_id` has to be provided.
* `schema_version_number` - (Required) Version number of the schema.

##### schema_id

* `registry_name` - (Optional) Name of the schema registry that contains the schema. Must be provided when `schema_name` is specified and conflicts with `schema_arn`.
* `schema_arn` - (Optional) ARN of the schema. One of `schema_arn` or `schema_name` has to be provided.
* `schema_name` - (Optional) Name of the schema. One of `schema_arn` or `schema_name` has to be provided.

#### ser_de_info

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) Map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is `org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe`.

#### sort_columns

* `column` - (Required) Name of the column.
* `sort_order` - (Required) Whether the column is sorted in ascending (`1`) or descending order (`0`).

#### skewed_info

* `skewed_column_names` - (Optional) List of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) List of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) Map of skewed values to the columns that contain them.

### target_table

* `catalog_id` - (Required) ID of the Data Catalog in which the table resides.
* `database_name` - (Required) Name of the catalog database that contains the target table.
* `name` - (Required) Name of the target table.
* `region` - (Optional) Region of the target table.

### view_definition

* `definer` - (Optional) The definer of a view in SQL.
* `is_protected` - (Optional) You can set this flag as true to instruct the engine not to push user-provided operations into the logical plan of the view during query planning. However, setting this flag does not guarantee that the engine will comply. Refer to the engine's documentation to understand the guarantees provided, if any.
* `last_refresh_type` - (Optional) Type of the materialized view's last refresh. Valid values: `Full`, `Incremental`.
* `refresh_seconds` - (Optional) Auto refresh interval in seconds for the materialized view.
* `representations` - (Optional) A list of structures that contains the dialect of the view, and the query that defines the view. See [`representations`](#representations) below.
* `sub_object_version_ids` - (Optional) List of the Apache Iceberg table versions referenced by the materialized view.
* `sub_objects` - (Optional) A list of base table ARNs that make up the view.
* `view_version_id` - (Optional) ID value that identifies this view's version. For materialized views, the version ID is the Apache Iceberg table's snapshot ID.
* `view_version_token` - (Optional) Version ID of the Apache Iceberg table.

#### representations

* `dialect` - (Optional) A parameter that specifies the engine type of a specific representation. Valid values are `REDSHIFT`, `ATHENA`, and `SPARK`.
* `dialect_version` - (Optional) A parameter that specifies the version of the engine of a specific representation.
* `validation_connection` - (Optional) The name of the connection to be used to validate the specific representation of the view.
* `view_expanded_text` - (Optional) A string that represents the SQL query that describes the view with expanded resource ARNs.
* `view_original_text` - (Optional) A string that represents the original SQL query that describes the view.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Glue Table.
* `id` - Catalog ID, Database name and of the name table.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Tables using the catalog ID (usually AWS account ID), database name, and table name. For example:

```terraform
import {
  to = aws_glue_catalog_table.MyTable
  id = "123456789012:MyDatabase:MyTable"
}
```

Using `terraform import`, import Glue Tables using the catalog ID (usually AWS account ID), database name, and table name. For example:

```console
% terraform import aws_glue_catalog_table.MyTable 123456789012:MyDatabase:MyTable
```
