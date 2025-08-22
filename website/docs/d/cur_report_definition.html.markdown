---
subcategory: "Cost and Usage Report"
layout: "aws"
page_title: "AWS: aws_cur_report_definition"
description: |-
  Get information on an AWS Cost and Usage Report Definition.
---

# Data Source: aws_cur_report_definition

Use this data source to get information on an AWS Cost and Usage Report Definition.

~> *NOTE:* The AWS Cost and Usage Report service is only available in `us-east-1` currently.

~> *NOTE:* If AWS Organizations is enabled, only the master account can use this resource.

## Example Usage

```terraform
data "aws_cur_report_definition" "report_definition" {
  report_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `report_name` - (Required) Name of the report definition to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `time_unit` - Frequency on which report data are measured and displayed.
* `format` - Preferred compression format for report.
* `compression` - Preferred format for report.
* `additional_schema_elements` - A list of schema elements.
* `s3_bucket` - Name of customer S3 bucket.
* `s3_prefix` - Preferred report path prefix.
* `s3_region` - Region of customer S3 bucket.
* `additional_artifacts` - A list of additional artifacts.
* `refresh_closed_reports` - If true reports are updated after they have been finalized.
* `report_versioning` - Overwrite the previous version of each report or to deliver the report in addition to the previous versions.
* `tags` - Map of key-value pairs assigned to the resource.
