---
subcategory: "BCM Dashboards"
layout: "aws"
page_title: "AWS: aws_bcmdashboards_scheduled_report"
description: |-
  Manages an AWS Billing and Cost Management Dashboards Scheduled Report.
---

# Resource: aws_bcmdashboards_scheduled_report

Manages an AWS Billing and Cost Management (BCM) Dashboards Scheduled Report. A scheduled report periodically delivers the contents of a dashboard.

## Example Usage

### Basic Usage

```terraform
resource "aws_bcmdashboards_scheduled_report" "example" {
  name                                = "example-report"
  dashboard_arn                       = aws_bcmdashboards_dashboard.example.arn
  scheduled_report_execution_role_arn = aws_iam_role.example.arn

  schedule_config {
    schedule_expression           = "cron(0 9 1 * ? *)"
    schedule_expression_time_zone = "UTC"
    state                         = "ENABLED"
  }
}
```

## Argument Reference

The following arguments are required:

* `dashboard_arn` - (Required) ARN of the dashboard to generate the scheduled report from.
* `name` - (Required) Name of the scheduled report.
* `schedule_config` - (Required) Configuration block defining when and how often the report is generated. See [`schedule_config`](#schedule_config-block) below.
* `scheduled_report_execution_role_arn` - (Required) ARN of the IAM role that the scheduled report assumes when executing.

The following arguments are optional:

* `description` - (Optional) Description of the scheduled report.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `widget_date_range_override` - (Optional) Configuration block specifying a date range override applied to the widgets in the scheduled report. See [`widget_date_range_override`](#widget_date_range_override-block) below.
* `widget_ids` - (Optional) List of widget identifiers to include in the scheduled report. If not specified, all widgets in the dashboard are included.

### `schedule_config` Block

* `schedule_expression` - (Optional) Schedule expression that specifies when to run the report. Must be a `cron` expression with six fields.
* `schedule_expression_time_zone` - (Optional) Time zone for the schedule expression, for example `UTC`.
* `schedule_period_end_time` - (Optional) End of the active period of the schedule, in RFC3339 format. Defaults to three years from creation if not specified.
* `schedule_period_start_time` - (Optional) Start of the active period of the schedule, in RFC3339 format. Defaults to the creation time if not specified.
* `state` - (Optional) State of the schedule. Valid values: `ENABLED`, `DISABLED`.

### `widget_date_range_override` Block

* `end_time` - (Required) Configuration block specifying the end of the date range. See [`end_time`](#widget_date_range_override-end_time-block) below.
* `start_time` - (Required) Configuration block specifying the start of the date range. See [`start_time`](#widget_date_range_override-start_time-block) below.

#### `widget_date_range_override` `start_time` Block

* `type` - (Required) Type of the date/time value. Valid values: `ABSOLUTE` or `RELATIVE`.
* `value` - (Required) Date/time value.

#### `widget_date_range_override` `end_time` Block

* `type` - (Required) Type of the date/time value. Valid values: `ABSOLUTE` or `RELATIVE`.
* `value` - (Required) Date/time value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the scheduled report.
* `created_at` - Timestamp when the scheduled report was created.
* `id` - (**Deprecated**) ARN of the scheduled report. Use `arn` instead.
* `last_execution_at` - Timestamp of the most recent execution of the scheduled report.
* `updated_at` - Timestamp when the scheduled report was last updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import BCM Dashboards Scheduled Report using the `arn`. For example:

```terraform
import {
  to = aws_bcmdashboards_scheduled_report.example
  id = "arn:aws:bcm-dashboards::123456789012:scheduled-report/01234567-89ab-cdef-0123-456789abcdef"
}
```

Using `terraform import`, import BCM Dashboards Scheduled Report using the `arn`. For example:

```console
% terraform import aws_bcmdashboards_scheduled_report.example arn:aws:bcm-dashboards::123456789012:scheduled-report/01234567-89ab-cdef-0123-456789abcdef
```
