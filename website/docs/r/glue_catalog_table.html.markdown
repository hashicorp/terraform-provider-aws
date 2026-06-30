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

* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.
* `name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.

The following arguments are optional:

* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `description` - (Optional) Description of the table.
* `open_table_format_input` - (Optional) Configuration block for open table formats. See [`open_table_format_input`](#open_table_format_input-block) below.
* `owner` - (Optional) Owner of the table.
* `parameters` - (Optional) Properties associated with this table, as a map of key-value pairs.
* `partition_index` - (Optional) Configuration block for a maximum of 3 partition indexes. See [`partition_index`](#partition_index-block) below.
* `partition_keys` - (Optional) Configuration block of columns by which the table is partitioned. Only primitive types are supported as partition keys. See [`partition_keys`](#partition_keys-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention` - (Optional) Retention time for this table.
* `storage_descriptor` - (Optional) Configuration block for information about the physical storage of this table. For more information, refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor). See [`storage_descriptor`](#storage_descriptor-block) below.
* `table_type` - (Optional) Type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
* `target_table` - (Optional) Configuration block of a target table for resource linking. See [`target_table`](#target_table-block) below.
* `view_definition` - (Optional) Structure that contains all the information that defines the view, including the dialect or dialects for the view, and the query. See [`view_definition`](#view_definition-block) below.
* `view_expanded_text` - (Optional) If the table is a view, the expanded text of the view; otherwise null.
* `view_original_text` - (Optional) If the table is a view, the original text of the view; otherwise null.

### `open_table_format_input` Block

~> **NOTE:** An `open_table_format_input` cannot be added to an existing `glue_catalog_table`.
This will destroy and recreate the table, possibly resulting in data loss.

* `iceberg_input` - (Required) Configuration block for iceberg table config. See [`iceberg_input`](#iceberg_input-block) below.

### `iceberg_input` Block

~> **NOTE:** An `iceberg_input` cannot be added to an existing `open_table_format_input`.
This will destroy and recreate the table, possibly resulting in data loss.

* `iceberg_table_input` - (Optional) Configuration parameters, including table properties and metadata specifications. See [`iceberg_table_input`](#iceberg_table_input-block) below.
* `metadata_operation` - (Required) Required metadata operation. Can only be set to CREATE.
* `version` - (Optional) Table version for the Iceberg table. Defaults to 2.

### `iceberg_table_input` Block

* `location` - (Required) S3 location where the Iceberg table data will be stored. Maximum length of 2056 characters.
* `partition_spec` - (Optional) Partitioning specification that defines how the Iceberg table data will be organized and partitioned for optimal query performance. See [`partition_spec`](#partition_spec-block) below.
* `properties` - (Optional) Key-value pairs of additional table properties and configuration settings for the Iceberg table.
* `schema` - (Required) Schema definition that specifies the structure, field types, and metadata for the Iceberg table. See [`schema`](#schema-block) below.
* `sort_order` - (Optional) Sort order specification that defines how data should be ordered within each partition to optimize query performance. See [`sort_order`](#sort_order-block) below.

### `partition_spec` Block

* `fields` - (Required) List of partition fields that define how the table data should be partitioned. See [`partition_spec.fields`](#partition_specfields-block) below.
* `spec_id` - (Optional) Unique identifier for this partition specification within the Iceberg table's metadata history.

#### `partition_spec.fields` Block

* `field_id` - (Optional) Unique identifier assigned to this partition field within the Iceberg table's partition specification.
* `name` - (Required) Name of the partition field as it will appear in the partitioned table structure. Length between 1 and 1024 characters.
* `source_id` - (Required) Identifier of the source field from the table schema that this partition field is based on.
* `transform` - (Required) Transformation function applied to the source field to create the partition. Common values: `identity`, `bucket`, `truncate`, `year`, `month`, `day`, `hour`.

### `schema` Block

* `fields` - (Required) List of field definitions that make up the table schema. See [`schema.fields`](#schemafields-block) below.
* `identifier_field_ids` - (Optional) List of field identifiers that uniquely identify records in the table, used for row-level operations and deduplication.
* `schema_id` - (Optional) Unique identifier for this schema version within the Iceberg table's schema evolution history.
* `type` - (Optional) Root type of the schema structure. Valid value: `struct`.

#### `schema.fields` Block

* `doc` - (Optional) Documentation or description text that provides additional context about the purpose and usage of this field. Length between 0 and 255 characters.
* `id` - (Required) Unique identifier assigned to this field within the Iceberg table schema, used for schema evolution and field tracking.
* `initial_default` - (Optional) Default value as JSON used to populate the field's value for all records that were written before the field was added to the schema.
* `name` - (Required) Name of the field as it appears in the table schema and query operations. Length between 1 and 1024 characters.
* `required` - (Required) Whether this field is required (non-nullable) or optional (nullable) in the table schema.
* `type` - (Required) Data type definition for this field as a JSON string, specifying the structure and format of the data it contains. Examples: `"long"`, `"string"`, `"timestamp"`, `"decimal(10,2)"`.
* `write_default` - (Optional) Default value as JSON used to populate the field's value for any records written after the field was added to the schema, if the writer does not supply the field's value.

### `sort_order` Block

* `fields` - (Required) List of fields and their sort directions that define the ordering criteria for the Iceberg table data. See [`sort_order.fields`](#sort_orderfields-block) below.
* `order_id` - (Required) Unique identifier for this sort order specification within the Iceberg table's metadata.

#### `sort_order.fields` Block

* `direction` - (Required) Sort direction for this field. Valid values: `asc`, `desc`.
* `null_order` - (Required) Ordering behavior for null values in this field. Valid values: `nulls-first`, `nulls-last`.
* `source_id` - (Required) Identifier of the source field from the table schema that this sort field is based on.
* `transform` - (Required) Transformation function applied to the source field before sorting. Common values: `identity`, `bucket`, `truncate`.

### `partition_index` Block

~> **NOTE:** A `partition_index` cannot be added to an existing `glue_catalog_table`.
This will destroy and recreate the table, possibly resulting in data loss.
To add an index to an existing table, see the [`glue_partition_index` resource](/docs/providers/aws/r/glue_partition_index.html) for configuration details.

* `index_name` - (Required) Name of the partition index.
* `keys` - (Required) Keys for the partition index.

### `partition_keys` Block

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Partition Key.
* `parameters` - (Optional) Map of key-value pairs.
* `type` - (Optional) Datatype of data in the Partition Key.

### `storage_descriptor` Block

* `additional_locations` - (Optional) List of locations that point to the path where a Delta table is located.
* `bucket_columns` - (Optional) List of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `columns` - (Optional) Configuration block for columns in the table. See [`columns`](#columns-block) below.
* `compressed` - (Optional) Whether the data in the table is compressed.
* `input_format` - (Optional) Input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `location` - (Optional) Physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `output_format` - (Optional) Output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `schema_reference` - (Optional) Object that references a schema stored in the AWS Glue Schema Registry. When creating a table, you can pass an empty list of columns for the schema, and instead use a schema reference. See [Schema Reference](#schema_reference-block) below.
* `ser_de_info` - (Optional) Configuration block for serialization and deserialization ("SerDe") information. See [`ser_de_info`](#ser_de_info-block) below.
* `skewed_info` - (Optional) Configuration block with information about values that appear very frequently in a column (skewed values). See [`skewed_info`](#skewed_info-block) below.
* `sort_columns` - (Optional) Configuration block for the sort order of each bucket in the table. See [`sort_columns`](#sort_columns-block) below.
* `stored_as_sub_directories` - (Optional) Whether the table data is stored in subdirectories.

#### `columns` Block

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Column.
* `parameters` - (Optional) Key-value pairs defining properties associated with the column.
* `type` - (Optional) Datatype of data in the Column.

#### `schema_reference` Block

* `schema_id` - (Optional) Configuration block that contains schema identity fields. Either this or the `schema_version_id` has to be provided. See [`schema_id`](#schema_id-block) below.
* `schema_version_id` - (Optional) Unique ID assigned to a version of the schema. Either this or the `schema_id` has to be provided.
* `schema_version_number` - (Required) Version number of the schema.

##### `schema_id` Block

* `registry_name` - (Optional) Name of the schema registry that contains the schema. Must be provided when `schema_name` is specified and conflicts with `schema_arn`.
* `schema_arn` - (Optional) ARN of the schema. One of `schema_arn` or `schema_name` has to be provided.
* `schema_name` - (Optional) Name of the schema. One of `schema_arn` or `schema_name` has to be provided.

#### `ser_de_info` Block

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) Map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is `org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe`.

#### `sort_columns` Block

* `column` - (Required) Name of the column.
* `sort_order` - (Required) Whether the column is sorted in ascending (`1`) or descending order (`0`).

#### `skewed_info` Block

* `skewed_column_names` - (Optional) List of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) List of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) Map of skewed values to the columns that contain them.

### `target_table` Block

* `catalog_id` - (Required) ID of the Data Catalog in which the table resides.
* `database_name` - (Required) Name of the catalog database that contains the target table.
* `name` - (Required) Name of the target table.
* `region` - (Optional) Region of the target table.

### `view_definition` Block

~> **NOTE:** Changes to a multi-dialect view (one configured via the `view_definition` block) are applied in place via Glue's [`UpdateTable`](https://docs.aws.amazon.com/glue/latest/webapi/API_UpdateTable.html) API with `ViewUpdateAction = REPLACE`. The table identity is preserved across the update, which preserves any [Lake Formation](https://docs.aws.amazon.com/lake-formation/latest/dg/access-control-overview.html) grants attached to the view. Legacy views configured only via the top-level `view_original_text` / `view_expanded_text` arguments are updated without `ViewUpdateAction`, as required by the Glue API.

* `definer` - (Optional) Definer of a view in SQL.
* `is_protected` - (Optional) You can set this flag as true to instruct the engine not to push user-provided operations into the logical plan of the view during query planning. However, setting this flag does not guarantee that the engine will comply. Refer to the engine's documentation to understand the guarantees provided, if any.
* `last_refresh_type` - (Optional) Type of the materialized view's last refresh. Valid values: `Full`, `Incremental`.
* `refresh_seconds` - (Optional) Auto refresh interval in seconds for the materialized view.
* `representations` - (Optional) List of structures that contains the dialect of the view, and the query that defines the view. See [`representations`](#representations-block) below.
* `sub_object_version_ids` - (Optional) List of the Apache Iceberg table versions referenced by the materialized view.
* `sub_objects` - (Optional) List of base table ARNs that make up the view.
* `view_version_id` - (Optional) ID value that identifies this view's version. For materialized views, the version ID is the Apache Iceberg table's snapshot ID.
* `view_version_token` - (Optional) Version ID of the Apache Iceberg table.

#### `representations` Block

* `dialect` - (Optional) Parameter that specifies the engine type of a specific representation. Valid values are `REDSHIFT`, `ATHENA`, and `SPARK`.
* `dialect_version` - (Optional) Parameter that specifies the version of the engine of a specific representation.
* `validation_connection` - (Optional) Name of the connection to be used to validate the specific representation of the view.
* `view_expanded_text` - (Optional) String that represents the SQL query that describes the view with expanded resource ARNs.
* `view_original_text` - (Optional) String that represents the original SQL query that describes the view.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Table.
* `id` - Catalog ID, database name, and table name, separated by colons (`:`).
* `partition_index[*].index_status` - Status of the partition index.

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
