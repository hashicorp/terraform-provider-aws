---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_partition_index"
description: |-
  Provides a Glue Partition Index.
---

# Resource: aws_glue_partition_index

## Example Usage

```terraform
resource "aws_glue_catalog_database" "example" {
  name = "example"
}

resource "aws_glue_catalog_table" "example" {
  name               = "example"
  database_name      = aws_glue_catalog_database.example.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}

resource "aws_glue_partition_index" "example" {
  database_name = aws_glue_catalog_database.example.name
  table_name    = aws_glue_catalog_table.example.name

  partition_index {
    index_name = "example"
    keys       = ["my_column_1", "my_column_2"]
  }
}
```

## Argument Reference

The following arguments are required:

* `table_name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.
* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.
* `partition_index` - (Required) Configuration block for a partition index. See [`partition_index`](#partition_index) below.
* `catalog_id` - (Optional) The catalog ID where the table resides.

### partition_index

* `index_name` - (Required) Name of the partition index.
* `keys` - (Required) Keys for the partition index.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Catalog ID, Database name, table name, and index name.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Partition Indexes using the catalog ID (usually AWS account ID), database name, table name, and index name. For example:

```terraform
import {
  to = aws_glue_partition_index.example
  id = "123456789012:MyDatabase:MyTable:index-name"
}
```

Using `terraform import`, import Glue Partition Indexes using the catalog ID (usually AWS account ID), database name, table name, and index name. For example:

```console
% terraform import aws_glue_partition_index.example 123456789012:MyDatabase:MyTable:index-name
```
