---
subcategory: "Cost and Usage Report"
layout: "aws"
page_title: "AWS: aws_cur_report_definition"
description: |-
  Provides a Cost and Usage Report Definition.
---

# Resource: aws_cur_report_definition

Manages Cost and Usage Report Definitions.

~> *NOTE:* The AWS Cost and Usage Report service is only available in `us-east-1` currently.

## Example Usage

```terraform
resource "aws_cur_report_definition" "example_cur_report_definition" {
  report_name                = "example-cur-report-definition"
  time_unit                  = "HOURLY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = "example-bucket-name"
  s3_region                  = "us-east-1"
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
```

## Argument Reference

This resource supports the following arguments:

* `report_name` - (Required) Unique name for the report. Must start with a number/letter and is case sensitive. Limited to 256 characters.
* `time_unit` - (Required) The frequency on which report data are measured and displayed.  Valid values are: `DAILY`, `HOURLY`, `MONTHLY`.
* `format` - (Required) Format for report. Valid values are: `textORcsv`, `Parquet`. If `Parquet` is used, then Compression must also be `Parquet`.
* `compression` - (Required) Compression format for report. Valid values are: `GZIP`, `ZIP`, `Parquet`. If `Parquet` is used, then format must also be `Parquet`.
* `additional_schema_elements` - (Required) A list of schema elements. Valid values are: `RESOURCES`, `SPLIT_COST_ALLOCATION_DATA`.
* `s3_bucket` - (Required) Name of the existing S3 bucket to hold generated reports.
* `s3_prefix` - (Optional) Report path prefix. Limited to 256 characters.
* `s3_region` - (Required) Region of the existing S3 bucket to hold generated reports.
* `additional_artifacts` - (Required) A list of additional artifacts. Valid values are: `REDSHIFT`, `QUICKSIGHT`, `ATHENA`. When ATHENA exists within additional_artifacts, no other artifact type can be declared and report_versioning must be `OVERWRITE_REPORT`.
* `refresh_closed_reports` - (Optional) Set to true to update your reports after they have been finalized if AWS detects charges related to previous months.
* `report_versioning` - (Optional) Overwrite the previous version of each report or to deliver the report in addition to the previous versions. Valid values are: `CREATE_NEW_REPORT` and `OVERWRITE_REPORT`.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) specifying the cur report.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Report Definitions using the `report_name`. For example:

```terraform
import {
  to = aws_cur_report_definition.example_cur_report_definition
  id = "example-cur-report-definition"
}
```

Using `terraform import`, import Report Definitions using the `report_name`. For example:

```console
% terraform import aws_cur_report_definition.example_cur_report_definition example-cur-report-definition
```
