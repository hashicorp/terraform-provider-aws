---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_report_plan"
description: |-
  Provides an AWS Backup Report Plan resource.
---

# Resource: aws_backup_report_plan

Provides an AWS Backup Report Plan resource.

## Example Usage

```terraform
resource "aws_backup_report_plan" "example" {
  name        = "example_name"
  description = "example description"

  report_delivery_channel {
    formats = [
      "CSV",
      "JSON"
    ]
    s3_bucket_name = "example-bucket-name"
  }

  report_setting {
    report_template = "RESTORE_JOB_REPORT"
  }

  tags = {
    "Name" = "Example Report Plan"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the report plan with a maximum of 1,024 characters
* `name` - (Required) The unique name of the report plan. The name must be between 1 and 256 characters, starting with a letter, and consisting of letters, numbers, and underscores.
* `report_delivery_channel` - (Required) An object that contains information about where and how to deliver your reports, specifically your Amazon S3 bucket name, S3 key prefix, and the formats of your reports. Detailed below.
* `report_setting` - (Required) An object that identifies the report template for the report. Reports are built using a report template. Detailed below.
* `tags` - (Optional) Metadata that you can assign to help organize the report plans you create. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Report Delivery Channel Arguments
For **report_delivery_channel** the following attributes are supported:

* `formats` - (Optional) A list of the format of your reports: CSV, JSON, or both. If not specified, the default format is CSV.
* `s3_bucket_name` - (Required) The unique name of the S3 bucket that receives your reports.
* `s3_key_prefix` - (Optional) The prefix for where Backup Audit Manager delivers your reports to Amazon S3. The prefix is this part of the following path: s3://your-bucket-name/prefix/Backup/us-west-2/year/month/day/report-name. If not specified, there is no prefix.

### Report Setting Arguments
For **report_setting** the following attributes are supported:

* `framework_arns` - (Optional) Specifies the Amazon Resource Names (ARNs) of the frameworks a report covers.
* `number_of_frameworks` - (Optional) Specifies the number of frameworks a report covers.
* `report_template` - (Required) Identifies the report template for the report. Reports are built using a report template. The report templates are: `RESOURCE_COMPLIANCE_REPORT` | `CONTROL_COMPLIANCE_REPORT` | `BACKUP_JOB_REPORT` | `COPY_JOB_REPORT` | `RESTORE_JOB_REPORT`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the backup report plan.
* `creation_time` - The date and time that a report plan is created, in Unix format and Coordinated Universal Time (UTC).
* `deployment_status` - The deployment status of a report plan. The statuses are: `CREATE_IN_PROGRESS` | `UPDATE_IN_PROGRESS` | `DELETE_IN_PROGRESS` | `COMPLETED`.
* `id` - The id of the backup report plan.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Backup Report Plan can be imported using the `id` which corresponds to the name of the Backup Report Plan, e.g.,

```
$ terraform import aws_backup_report_plan.test <id>
```
