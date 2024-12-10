---
subcategory: "BCM Data Exports"
layout: "aws"
page_title: "AWS: aws_bcmdataexports_export"
description: |-
  Terraform resource for managing an AWS BCM Data Exports Export.
---

# Resource: aws_bcmdataexports_export

Terraform resource for managing an AWS BCM Data Exports Export.

## Example Usage

### Basic Usage

```terraform
resource "aws_bcmdataexports_export" "test" {
  export {
    name = "testexample"
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        COST_AND_USAGE_REPORT = {
          TIME_GRANULARITY                      = "HOURLY",
          INCLUDE_RESOURCES                     = "FALSE",
          INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY = "FALSE",
          INCLUDE_SPLIT_COST_ALLOCATION_DATA    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }

    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `export` - (Required) The details of the export, including data query, name, description, and destination configuration.  See the [`export` argument reference](#export-argument-reference) below.

### `export` Argument Reference

* `data_query` - (Required) Data query for this specific data export. See the [`data_query` argument reference](#data_query-argument-reference) below.
* `destination_configurations` - (Required) Destination configuration for this specific data export. See the [`destination_configurations` argument reference](#destination_configurations-argument-reference) below.
* `name` - (Required) Name of this specific data export.
* `refresh_cadence` - (Required) Cadence for Amazon Web Services to update the export in your S3 bucket. See the [`refresh_cadence` argument reference](#refresh_cadence-argument-reference) below.
* `description` - (Optional) Description for this specific data export.

### `data_query` Argument Reference

* `query_statement` - (Required) Query statement.
* `table_configurations` - (Optional) Table configuration.

### `destination_configurations` Argument Reference

* `s3_destination` - (Required) Object that describes the destination of the data exports file. See the [`s3_destination` argument reference](#s3_destination-argument-reference) below.

### `s3_destination` Argument Reference

* `s3_bucket` - (Required) Name of the Amazon S3 bucket used as the destination of a data export file.
* `s3_output_configurations` - (Required) Output configuration for the data export. See the [`s3_output_configurations` argument reference](#s3_output_configurations-argument-reference) below.
* `s3_prefix` - (Required) S3 path prefix you want prepended to the name of your data export.
* `s3_region` - (Required) S3 bucket region.

### `s3_output_configurations` Argument Reference

* `compression` - (Required) Compression type for the data export. Valid values `GZIP`, `PARQUET`.
* `format` - (Required) File format for the data export. Valid values `TEXT_OR_CSV` or `PARQUET`.
* `output_type` - (Required) Output type for the data export. Valid value `CUSTOM`.
* `overwrite` - (Required) The rule to follow when generating a version of the data export file. You have the choice to overwrite the previous version or to be delivered in addition to the previous versions. Overwriting exports can save on Amazon S3 storage costs. Creating new export versions allows you to track the changes in cost and usage data over time. Valid values `CREATE_NEW_REPORT` or `OVERWRITE_REPORT`.

### `refresh_cadence` Argument Reference

* `frequency` - (Required) Frequency that data exports are updated. The export refreshes each time the source data updates, up to three times daily. Valid values `SYNCHRONOUS`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `export_arn` - Amazon Resource Name (ARN) for this export.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import BCM Data Exports Export using the export ARN. For example:

```terraform
import {
  to = aws_bcmdataexports_export.example
  id = "arn:aws:bcm-data-exports:us-east-1:123456789012:export/CostUsageReport-9f1c75f3-f982-4d9a-b936-1e7ecab814b7"
}
```

Using `terraform import`, import BCM Data Exports Export using the export ARN. For example:

```console
% terraform import aws_bcmdataexports_export.example arn:aws:bcm-data-exports:us-east-1:123456789012:export/CostUsageReport-9f1c75f3-f982-4d9a-b936-1e7ecab814b7
```
