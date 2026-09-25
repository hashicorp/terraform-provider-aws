---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_set"
description: |-
  Manages a Resource QuickSight Data Set.
---

# Resource: aws_quicksight_data_set

Resource for managing a QuickSight Data Set.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
}
```

### With use_as

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"
  use_as      = "RLS_RULES"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "UserName"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
}
```

### With Column Level Permission Rules

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
  column_level_permission_rules {
    column_names = ["Column1"]
    principals   = [aws_quicksight_user.example.arn]
  }
}
```

### With Field Folders

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
  field_folders {
    field_folders_id = "example-id"
    columns          = ["Column1"]
    description      = "example description"
  }
}
```

### With Permissions

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
  permissions {
    actions = [
      "quicksight:DescribeDataSet",
      "quicksight:DescribeDataSetPermissions",
      "quicksight:PassDataSet",
      "quicksight:DescribeIngestion",
      "quicksight:ListIngestions",
    ]
    principal = aws_quicksight_user.example.arn
  }
}
```

### With Row Level Permission Tag Configuration

```terraform
resource "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
  name        = "example-name"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "example-id"
    s3_source {
      data_source_arn = aws_quicksight_data_source.example.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
  row_level_permission_tag_configuration {
    status = "ENABLED"
    tag_rules {
      column_name               = "Column1"
      tag_key                   = "tagkey"
      match_all_value           = "*"
      tag_multi_value_delimiter = ","
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required, Forces new resource) Identifier for the data set.
* `import_mode` - (Required) Whether to import the data into SPICE. Valid values are `SPICE` and `DIRECT_QUERY`.
* `name` - (Required) Display name for the dataset.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `column_groups` - (Optional) Groupings of columns that work together in certain Amazon QuickSight features. Currently, only geospatial hierarchy is supported. See [`column_groups` Block](#column_groups-block) below.
* `column_level_permission_rules` - (Optional) Set of 1 or more definitions of a [ColumnLevelPermissionRule](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnLevelPermissionRule.html). See [`column_level_permission_rules` Block](#column_level_permission_rules-block) below.
* `data_set_usage_configuration` - (Optional) Usage configuration to apply to child datasets that reference this dataset as a source. See [`data_set_usage_configuration` Block](#data_set_usage_configuration-block) below.
* `field_folders` - (Optional) Folder that contains fields and nested subfolders for your dataset. See [`field_folders` Block](#field_folders-block) below.
* `logical_table_map` - (Optional) Configures the combination and transformation of the data from the physical tables. Maximum of 1 entry. See [`logical_table_map` Block](#logical_table_map-block) below.
* `permissions` - (Optional) Set of resource permissions on the data source. Maximum of 64 items. See [`permissions` Block](#permissions-block) below.
* `physical_table_map` - (Optional) Declares the physical tables that are available in the underlying data sources. See [`physical_table_map` Block](#physical_table_map-block) below.
* `refresh_properties` - (Optional) Refresh properties for the data set. **NOTE**: Only valid when `import_mode` is set to `SPICE`. See [`refresh_properties` Block](#refresh_properties-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `row_level_permission_data_set` - (Optional) Row-level security configuration for the data that you want to create. See [`row_level_permission_data_set` Block](#row_level_permission_data_set-block) below.
* `row_level_permission_tag_configuration` - (Optional) Configuration of tags on a dataset to set row-level security. Row-level security tags are currently supported for anonymous embedding only. See [`row_level_permission_tag_configuration` Block](#row_level_permission_tag_configuration-block) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `use_as` - (Optional, Forces new resource) Purpose of the data set. The only valid value is `RLS_RULES`, which designates this data set as a Row Level Security (RLS) rules dataset. An RLS rules dataset is used to control access to data at the row level in QuickSight analyses and dashboards. See the [AWS documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CreateDataSet.html#API_CreateDataSet_RequestSyntax) for details.

### `physical_table_map` Block

For a `physical_table_map` item to be valid, only one of `custom_sql`, `relational_table`, or `s3_source` should be configured.

* `custom_sql` - (Optional) Physical table type built from the results of the custom SQL query. See [`custom_sql` Block](#custom_sql-block) below.
* `physical_table_map_id` - (Required) Key of the physical table map.
* `relational_table` - (Optional) Physical table type for relational data sources. See [`relational_table` Block](#relational_table-block) below.
* `s3_source` - (Optional) Physical table type for an S3 data source. See [`s3_source` Block](#s3_source-block) below.

### `custom_sql` Block

* `columns` - (Optional) Column schema from the SQL query result set. See [`physical_table_map.custom_sql.columns` Block](#physical_table_mapcustom_sqlcolumns-block) below.
* `data_source_arn` - (Required) ARN of the data source.
* `name` - (Required) Display name for the SQL query result.
* `sql_query` - (Required) SQL query.

### `physical_table_map.custom_sql.columns` Block

* `name` - (Required) Name of this column in the underlying data source.
* `type` - (Required) Data type of the column.

### `relational_table` Block

* `catalog` - (Optional) Catalog associated with the table.
* `data_source_arn` - (Required) ARN of the data source.
* `input_columns` - (Required) Column schema of the table. See [`input_columns` Block](#input_columns-block) below.
* `name` - (Required) Name of the relational table.
* `schema` - (Optional) Schema name. This name applies to certain relational database engines.

### `input_columns` Block

* `name` - (Required) Name of this column in the underlying data source.
* `type` - (Required) Data type of the column.

### `s3_source` Block

* `data_source_arn` - (Required) ARN of the data source.
* `input_columns` - (Required) Column schema of the table. See [`input_columns` Block](#input_columns-block) below.
* `upload_settings` - (Required) Information about the format for the S3 source file or files. See [`upload_settings` Block](#upload_settings-block) below.

### `upload_settings` Block

* `contains_header` - (Optional) Whether the file has a header row, or the files each have a header row.
* `delimiter` - (Optional) Delimiter between values in the file.
* `format` - (Optional) File format. Valid values are `CSV`, `TSV`, `CLF`, `ELF`, `XLSX`, and `JSON`.
* `start_from_row` - (Optional) Row number to start reading data from.
* `text_qualifier` - (Optional) Text qualifier. Valid values are `DOUBLE_QUOTE` and `SINGLE_QUOTE`.

### `column_groups` Block

* `geo_spatial_column_group` - (Optional) Geospatial column group that denotes a hierarchy. See [`geo_spatial_column_group` Block](#geo_spatial_column_group-block) below.

### `geo_spatial_column_group` Block

* `columns` - (Required) Columns in this hierarchy.
* `country_code` - (Required) Country code. Valid values are `US`.
* `name` - (Required) Display name for the hierarchy.

### `column_level_permission_rules` Block

* `column_names` - (Optional) Array of column names.
* `principals` - (Optional) Array of ARNs for Amazon QuickSight users or groups.

### `data_set_usage_configuration` Block

* `disable_use_as_direct_query_source` - (Optional) Controls whether a child dataset of a direct query can use this dataset as a source.
* `disable_use_as_imported_source` - (Optional) Controls whether a child dataset that's stored in QuickSight can use this dataset as a source.

### `field_folders` Block

* `columns` - (Optional) Array of column names to add to the folder. A column can only be in one folder.
* `description` - (Optional) Field folder description.
* `field_folders_id` - (Required) Key of the field folder map.

### `logical_table_map` Block

* `alias` - (Required) Display name for the logical table.
* `data_transforms` - (Optional) Transform operations that act on this logical table. For this structure to be valid, only one of the attributes can be non-null. See [`data_transforms` Block](#data_transforms-block) below.
* `logical_table_map_id` - (Required) Key of the logical table map.
* `source` - (Optional) Source of this logical table. See [`source` Block](#source-block) below.

### `data_transforms` Block

* `cast_column_type_operation` - (Optional) Transform operation that casts a column to a different type. See [`cast_column_type_operation` Block](#cast_column_type_operation-block) below.
* `create_columns_operation` - (Optional) Operation that creates calculated columns. Columns created in one such operation form a lexical closure. See [`create_columns_operation` Block](#create_columns_operation-block) below.
* `filter_operation` - (Optional) Operation that filters rows based on some condition. See [`filter_operation` Block](#filter_operation-block) below.
* `project_operation` - (Optional) Operation that projects columns. Operations that come after a projection can only refer to projected columns. See [`project_operation` Block](#project_operation-block) below.
* `rename_column_operation` - (Optional) Operation that renames a column. See [`rename_column_operation` Block](#rename_column_operation-block) below.
* `tag_column_operation` - (Optional) Operation that tags a column with additional information. See [`tag_column_operation` Block](#tag_column_operation-block) below.
* `untag_column_operation` - (Optional) Transform operation that removes tags associated with a column. See [`untag_column_operation` Block](#untag_column_operation-block) below.

### `cast_column_type_operation` Block

* `column_name` - (Required) Column name.
* `format` - (Optional) When casting a column from string to datetime type, you can supply a string in a format supported by Amazon QuickSight to denote the source data format.
* `new_column_type` - (Required) New column data type. Valid values are `STRING`, `INTEGER`, `DECIMAL`, `DATETIME`.

### `create_columns_operation` Block

* `columns` - (Required) Calculated columns to create. See [`logical_table_map.data_transforms.create_columns_operation.columns` Block](#logical_table_mapdata_transformscreate_columns_operationcolumns-block) below.

### `logical_table_map.data_transforms.create_columns_operation.columns` Block

* `column_id` - (Required) Unique ID to identify a calculated column. During a dataset update, if the column ID of a calculated column matches that of an existing calculated column, Amazon QuickSight preserves the existing calculated column.
* `column_name` - (Required) Column name.
* `expression` - (Required) Expression that defines the calculated column.

### `filter_operation` Block

* `condition_expression` - (Required) Expression that must evaluate to a Boolean value. Rows for which the expression evaluates to true are kept in the dataset.

### `project_operation` Block

* `projected_columns` - (Required) Projected columns.

### `rename_column_operation` Block

* `column_name` - (Required) Column to be renamed.
* `new_column_name` - (Required) New name for the column.

### `tag_column_operation` Block

* `column_name` - (Required) Column name.
* `tags` - (Required) Dataset column tag, currently only used for geospatial type tagging. See [`tags` Block](#tags-block) below.

### `tags` Block

* `column_description` - (Optional) Description for a column. See [`column_description` Block](#column_description-block) below.
* `column_geographic_role` - (Optional) Geospatial role for a column. Valid values are `COUNTRY`, `STATE`, `COUNTY`, `CITY`, `POSTCODE`, `LONGITUDE`, and `LATITUDE`.

### `column_description` Block

* `text` - (Optional) Text of a description for a column.

### `untag_column_operation` Block

* `column_name` - (Required) Column name.
* `tag_names` - (Required) Column tags to remove from this column.

### `source` Block

* `data_set_arn` - (Optional) ARN of the parent data set.
* `join_instruction` - (Optional) Result of a join of two logical tables. See [`join_instruction` Block](#join_instruction-block) below.
* `physical_table_id` - (Optional) Physical table ID.

### `join_instruction` Block

* `left_join_key_properties` - (Optional) Join key properties of the left operand. See [`left_join_key_properties` Block](#left_join_key_properties-block) below.
* `left_operand` - (Required) Operand on the left side of a join.
* `on_clause` - (Required) Join instructions provided in the ON clause of a join.
* `right_join_key_properties` - (Optional) Join key properties of the right operand. See [`right_join_key_properties` Block](#right_join_key_properties-block) below.
* `right_operand` - (Required) Operand on the right side of a join.
* `type` - (Required) Type of join. Valid values are `INNER`, `OUTER`, `LEFT`, and `RIGHT`.

### `left_join_key_properties` Block

* `unique_key` - (Optional) Value that indicates that a row in a table is uniquely identified by the columns in a join key. This is used by Amazon QuickSight to optimize query performance.

### `right_join_key_properties` Block

* `unique_key` - (Optional) Value that indicates that a row in a table is uniquely identified by the columns in a join key. This is used by Amazon QuickSight to optimize query performance.

### `permissions` Block

* `actions` - (Required) List of IAM actions to grant or revoke permissions on.
* `principal` - (Required) ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

### `row_level_permission_data_set` Block

* `arn` - (Required) ARN of the dataset that contains permissions for RLS.
* `format_version` - (Optional) User or group rules associated with the dataset that contains permissions for RLS.
* `namespace` - (Optional) Namespace associated with the dataset that contains permissions for RLS.
* `permission_policy` - (Required) Type of permissions to use when interpreting the permissions for RLS. Valid values are `GRANT_ACCESS` and `DENY_ACCESS`.
* `status` - (Optional) Status of the row-level security permission dataset. If enabled, the status is `ENABLED`. If disabled, the status is `DISABLED`.

### `row_level_permission_tag_configuration` Block

* `status` - (Optional) Status of row-level security tags. If enabled, the status is `ENABLED`. If disabled, the status is `DISABLED`.
* `tag_rules` - (Required) Set of rules associated with row-level security, such as the tag names and columns that they are assigned to. See [`tag_rules` Block](#tag_rules-block) below.

### `refresh_properties` Block

* `refresh_configuration` - (Required) Refresh configuration for the data set. See [`refresh_configuration` Block](#refresh_configuration-block) below.

### `refresh_configuration` Block

* `incremental_refresh` - (Required) Incremental refresh for the data set. See [`incremental_refresh` Block](#incremental_refresh-block) below.

### `incremental_refresh` Block

* `lookback_window` - (Required) Lookback window setup for an incremental refresh configuration. See [`lookback_window` Block](#lookback_window-block) below.

### `lookback_window` Block

* `column_name` - (Required) Name of the lookback window column.
* `size` - (Required) Lookback window column size.
* `size_unit` - (Required) Size unit that is used for the lookback window column. Valid values for this structure are `HOUR`, `DAY`, and `WEEK`.

### `tag_rules` Block

* `column_name` - (Required) Column name that a tag key is assigned to.
* `match_all_value` - (Optional) String that you want to use to filter by all the values in a column in the dataset and don’t want to list the values one by one.
* `tag_key` - (Required) Unique key for a tag.
* `tag_multi_value_delimiter` - (Optional) String that you want to use to delimit the values when you pass the values at run time.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the data set.
* `id` - Comma-delimited string joining AWS account ID and data set ID.
* `output_columns` - Final set of columns available for use in analyses and dashboards after all data preparation and transformation steps have been applied within the data set. See [`output_columns` Block](#output_columns-block) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

### `output_columns` Block

The `output_columns` block has the following attributes.

* `description` - Description of the column.
* `name` - Name of the column.
* `type` - Data type of the column.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Data Set using the AWS account ID and data set ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_data_set.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight Data Set using the AWS account ID and data set ID separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_data_set.example 123456789012,example-id
```
