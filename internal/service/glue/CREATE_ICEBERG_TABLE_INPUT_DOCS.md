# Glue CreateIcebergTableInput Documentation

This document describes the Terraform schema for AWS Glue `CreateIcebergTableInput`.

## Example Usage

### Basic Iceberg Table

```terraform
resource "aws_glue_catalog_table" "iceberg_table" {
  name          = "my_iceberg_table"
  database_name = "my_database"

  create_iceberg_table_input {
    location = "s3://my-bucket/iceberg-tables/my-table"

    schema {
      fields {
        id       = 1
        name     = "user_id"
        required = true
        type     = "\"long\""
        doc      = "Unique user identifier"
      }

      fields {
        id       = 2
        name     = "username"
        required = true
        type     = "\"string\""
      }

      fields {
        id       = 3
        name     = "created_at"
        required = false
        type     = "\"timestamp\""
      }
    }
  }
}
```

### Iceberg Table with Partitioning

```terraform
resource "aws_glue_catalog_table" "partitioned_iceberg_table" {
  name          = "my_partitioned_table"
  database_name = "my_database"

  create_iceberg_table_input {
    location = "s3://my-bucket/iceberg-tables/partitioned-table"

    schema {
      fields {
        id       = 1
        name     = "event_id"
        required = true
        type     = "\"long\""
      }

      fields {
        id       = 2
        name     = "event_time"
        required = true
        type     = "\"timestamp\""
      }

      fields {
        id       = 3
        name     = "event_type"
        required = true
        type     = "\"string\""
      }
    }

    partition_spec {
      fields {
        name      = "event_year"
        source_id = 2
        transform = "year"
      }

      fields {
        name      = "event_month"
        source_id = 2
        transform = "month"
      }
    }
  }
}
```

### Iceberg Table with Sort Order

```terraform
resource "aws_glue_catalog_table" "sorted_iceberg_table" {
  name          = "my_sorted_table"
  database_name = "my_database"

  create_iceberg_table_input {
    location = "s3://my-bucket/iceberg-tables/sorted-table"

    schema {
      fields {
        id       = 1
        name     = "transaction_id"
        required = true
        type     = "\"long\""
      }

      fields {
        id       = 2
        name     = "amount"
        required = true
        type     = "\"decimal(10,2)\""
      }

      fields {
        id       = 3
        name     = "timestamp"
        required = true
        type     = "\"timestamp\""
      }
    }

    write_order {
      order_id = 1

      fields {
        direction  = "desc"
        null_order = "nulls-last"
        source_id  = 3
        transform  = "identity"
      }

      fields {
        direction  = "asc"
        null_order = "nulls-first"
        source_id  = 1
        transform  = "identity"
      }
    }
  }
}
```

### Complete Iceberg Table Configuration

```terraform
resource "aws_glue_catalog_table" "complete_iceberg_table" {
  name          = "my_complete_table"
  database_name = "my_database"

  create_iceberg_table_input {
    location = "s3://my-bucket/iceberg-tables/complete-table"

    properties = {
      "write.format.default"       = "parquet"
      "write.metadata.compression" = "gzip"
      "write.parquet.compression"  = "snappy"
    }

    schema {
      fields {
        doc      = "Primary key"
        id       = 1
        name     = "id"
        required = true
        type     = "\"long\""
      }

      fields {
        doc           = "Record status"
        id            = 2
        name          = "status"
        required      = false
        type          = "\"string\""
        write_default = "\"active\""
      }

      fields {
        id       = 3
        name     = "created_date"
        required = true
        type     = "\"date\""
      }

      identifier_field_ids = [1]
      schema_id            = 0
      type                 = "struct"
    }

    partition_spec {
      fields {
        field_id  = 1000
        name      = "created_year"
        source_id = 3
        transform = "year"
      }

      fields {
        field_id  = 1001
        name      = "created_month"
        source_id = 3
        transform = "month"
      }

      spec_id = 0
    }

    write_order {
      fields {
        direction  = "desc"
        null_order = "nulls-last"
        source_id  = 3
        transform  = "identity"
      }

      order_id = 1
    }
  }
}
```

## Argument Reference

### create_iceberg_table_input

* `location` - (Required) The S3 location where the Iceberg table data will be stored. Maximum length of 2056 characters.
* `partition_spec` - (Optional) The partitioning specification that defines how the Iceberg table data will be organized and partitioned for optimal query performance. See [`partition_spec`](#partition_spec) below.
* `properties` - (Optional) Key-value pairs of additional table properties and configuration settings for the Iceberg table.
* `schema` - (Required) The schema definition that specifies the structure, field types, and metadata for the Iceberg table. See [`schema`](#schema) below.
* `write_order` - (Optional) The sort order specification that defines how data should be ordered within each partition to optimize query performance. See [`write_order`](#write_order) below.

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

### write_order

* `fields` - (Required) The list of fields and their sort directions that define the ordering criteria for the Iceberg table data. See [`fields`](#sort-fields) below.
* `order_id` - (Required) The unique identifier for this sort order specification within the Iceberg table's metadata.

#### sort fields

* `direction` - (Required) The sort direction for this field. Valid values: `asc`, `desc`.
* `null_order` - (Required) The ordering behavior for null values in this field. Valid values: `nulls-first`, `nulls-last`.
* `source_id` - (Required) The identifier of the source field from the table schema that this sort field is based on.
* `transform` - (Required) The transformation function applied to the source field before sorting. Common values: `identity`, `bucket`, `truncate`.

## Notes

* The `type` field in schema fields must be a valid JSON string representing an Iceberg data type.
* Field IDs in the schema must be unique within the table.
* Partition transforms must be compatible with the source field's data type.
* Sort order transforms must be compatible with the source field's data type.
