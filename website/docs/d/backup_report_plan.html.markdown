---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_report_plan"
description: |-
  Provides details about an AWS Backup Report Plan.
---

# Data Source: aws_backup_report_plan

Use this data source to get information on an existing backup report plan.

## Example Usage

```terraform
data "aws_backup_report_plan" "example" {
  name = "tf_example_backup_report_plan_name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The backup report plan name.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - The ARN of the backup report plan.
* `creation_time` - The date and time that a report plan is created, in Unix format and Coordinated Universal Time (UTC).
* `deployment_status` - The deployment status of a report plan. The statuses are: `CREATE_IN_PROGRESS` | `UPDATE_IN_PROGRESS` | `DELETE_IN_PROGRESS` | `COMPLETED`.
* `description` - The description of the report plan.
* `id` - The id of the report plan.
* `report_delivery_channel` - An object that contains information about where and how to deliver your reports, specifically your Amazon S3 bucket name, S3 key prefix, and the formats of your reports. Detailed below.
* `report_setting` - An object that identifies the report template for the report. Reports are built using a report template. Detailed below.
* `tags` - Metadata that you can assign to help organize the report plans you create.

### Report Delivery Channel Attributes
For **report_delivery_channel** the following attributes are supported:

* `formats` - A list of the format of your reports: CSV, JSON, or both.
* `s3_bucket_name` - The unique name of the S3 bucket that receives your reports.
* `s3_key_prefix` - The prefix for where Backup Audit Manager delivers your reports to Amazon S3. The prefix is this part of the following path: s3://your-bucket-name/prefix/Backup/us-west-2/year/month/day/report-name.

### Report Setting Attributes
For **report_setting** the following attributes are supported:

* `framework_arns` - Specifies the Amazon Resource Names (ARNs) of the frameworks a report covers.
* `number_of_frameworks` - Specifies the number of frameworks a report covers.
* `report_template` - Identifies the report template for the report. Reports are built using a report template.
