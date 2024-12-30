---
subcategory: "Glue DataBrew"
layout: "aws"
page_title: "AWS: aws_databrew_dataset"
description: |-
  Terraform resource for managing an AWS DataBrew Dataset.
---

# Resource: aws_databrew_dataset

Terraform resource for managing an AWS DataBrew Dataset.

## Example Usage

### Basic Usage

```terraform
resource "aws_databrew_dataset" "example" {
  name = "test"
  input {
    s3_input_definition {
      bucket = "bucket-name"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the dataset to be created.

* `input` - (Required) Represents information on how DataBrew can find data, in either the AWS Glue Data Catalog or Amazon S3.

* `format` - (Optional) The file format of a dataset that is created from an Amazon S3 file or folder.

* `format_options` - (Optional) Represents a set of options that define the structure of either comma-separated value (CSV), Excel, or JSON input.

* `path_options` - (Optional) A set of options that defines how DataBrew interprets an Amazon S3 path of the dataset.

### input

Supported nested arguments for the `input` configuration block:

* `database_input_definition` - (Optional) Connection information for dataset input files stored in a database.
    * `glue_connection_name` (Required) - The AWS Glue Connection that stores the connection information for the target database.
    * `database_table_name` (Optional) - The table within the target database.
    * `query_string` (Optional) - Custom SQL to run against the provided AWS Glue connection. This SQL will be used as the input for DataBrew projects and jobs.
    * `temp_directory` (Optional) - Represents an Amazon S3 location (bucket name, bucket owner, and object key) where DataBrew can read input data, or write output from a job.
        * `bucket` (Required) - The Amazon S3 bucket name.
        * `bucket_owner` (Optional) - The AWS account ID of the bucket owner.
        * `key` (Optional) - The unique name of the object in the bucket.

* `data_catalog_input_definition` - (Optional) The AWS Glue Data Catalog parameters for the data.
    * `database_name` (Required) - The name of a database in the Data Catalog.
    * `table_name` (Required) - The name of a database table in the Data Catalog. This table corresponds to a DataBrew dataset.
    * `catalog_id` (Optional) - The unique identifier of the AWS account that holds the Data Catalog that stores the data.
    * `temp_directory` (Optional) - Represents an Amazon S3 location (bucket name, bucket owner, and object key) where DataBrew can read input data, or write output from a job.
        * `bucket` (Required) - The Amazon S3 bucket name.
        * `bucket_owner` (Optional) - The AWS account ID of the bucket owner.
        * `key` (Optional) - The unique name of the object in the bucket.

* `metadata` - (Optional) Contains additional resource information needed for specific datasets.
    * `source_arn` (Optional) - The Amazon Resource Name (ARN) associated with the dataset. Currently, DataBrew only supports ARNs from Amazon AppFlow.

* `s3_input_definition` - (Optional) The Amazon S3 location where the data is stored.
    * `bucket` (Required) - The Amazon S3 bucket name.
    * `bucket_owner` (Optional) - The AWS account ID of the bucket owner.
    * `key` (Optional) - The unique name of the object in the bucket.

### format_options

Supported nested arguments for the `format_options` configuration block:

* `csv` - (Optional) Options that define how CSV input is to be interpreted by DataBrew.
    * `delimiter` (Optional) - A single character that specifies the delimiter being used in the CSV file.
    * `header_row` (Optional) - A variable that specifies whether the first row in the file is parsed as the header. If this value is false, column names are auto-generated.

* `excel` - (Optional) Options that define how Excel input is to be interpreted by DataBrew.
    * `header_row` (Optional) - A variable that specifies whether the first row in the file is parsed as the header. If this value is false, column names are auto-generated.
    * `sheet_indexes` (Optional) - One or more sheet numbers in the Excel file that will be included in the dataset.
    * `sheet_names` (Optional) - One or more named sheets in the Excel file that will be included in the dataset.

* `json` - (Optional) Options that define how JSON input is to be interpreted by DataBrew.
    * `multi_line` (Optional) - Represents the JSON-specific options that define how input is to be interpreted by AWS Glue DataBrew.

### path_options

Supported nested arguments for the `path_options` configuration block:

* `files_limit` - (Optional) If provided, this structure imposes a limit on a number of files that should be selected.
    * `max_files` (Required) - The number of Amazon S3 files to select.
    * `order` (Optional) - A criteria to use for Amazon S3 files sorting before their selection. By default uses DESCENDING order, i.e. most recent files are selected first. Another possible value is ASCENDING.
    * `ordered_by` (Optional) - A criteria to use for Amazon S3 files sorting before their selection. By default uses LAST_MODIFIED_DATE as a sorting criteria. Currently it's the only allowed value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the Dataset (same as name).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataBrew Project using the `example_id_arg`. For example:

```terraform
import {
  to = aws_databrew_dataset.example
  id = "project-dataset-name"
}
```

Using `terraform import`, import DataBrew Dataset using the `example_id_arg`. For example:

```console
% terraform import aws_databrew_dataset.example dataset-name
```
