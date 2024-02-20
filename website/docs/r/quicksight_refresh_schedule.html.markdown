---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_refresh_schedule"
description: |-
  Manages a Resource QuickSight Refresh Schedule.
---

# Resource: aws_quicksight_refresh_schedule

Resource for managing a QuickSight Refresh Schedule.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_refresh_schedule" "example" {
  data_set_id = "dataset-id"
  schedule_id = "schedule-id"

  schedule {
    refresh_type = "FULL_REFRESH"

    schedule_frequency {
      interval = "HOURLY"
    }
  }
}
```

### With Weekly Refresh

```terraform
resource "aws_quicksight_refresh_schedule" "example" {
  data_set_id = "dataset-id"
  schedule_id = "schedule-id"

  schedule {
    refresh_type = "INCREMENTAL_REFRESH"

    schedule_frequency {
      interval        = "WEEKLY"
      time_of_the_day = "01:00"
      timezone        = "Europe/London"
      refresh_on_day {
        day_of_week = "MONDAY"
      }
    }
  }
}
```

### With Monthly Refresh

```terraform
resource "aws_quicksight_refresh_schedule" "example" {
  data_set_id = "dataset-id"
  schedule_id = "schedule-id"

  schedule {
    refresh_type = "INCREMENTAL_REFRESH"

    schedule_frequency {
      interval        = "MONTHLY"
      time_of_the_day = "01:00"
      timezone        = "Europe/London"
      refresh_on_day {
        day_of_month = "1"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required, Forces new resource) The ID of the dataset.
* `schedule_id` - (Required, Forces new resource) The ID of the refresh schedule.
* `schedule` - (Required) The [refresh schedule](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RefreshSchedule.html). See [schedule](#schedule)

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID.

### schedule

* `refresh_type` - (Required) The type of refresh that the dataset undergoes. Valid values are `INCREMENTAL_REFRESH` and `FULL_REFRESH`.
* `start_after_date_time` (Optional) Time after which the refresh schedule can be started, expressed in `YYYY-MM-DDTHH:MM:SS` format.
* `schedule_frequency` - (Optional) The configuration of the [schedule frequency](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RefreshFrequency.html). See [schedule_frequency](#schedule_frequency).

### schedule_frequency

* `interval` - (Required) The interval between scheduled refreshes. Valid values are `MINUTE15`, `MINUTE30`, `HOURLY`, `DAILY`, `WEEKLY` and `MONTHLY`.
* `time_of_the_day` - (Optional) The time of day that you want the dataset to refresh. This value is expressed in `HH:MM` format. This field is not required for schedules that refresh hourly.
* `timezone` - (Optional) The timezone that you want the refresh schedule to use.
* `refresh_on_day` - (Optional) The [refresh on entity](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScheduleRefreshOnEntity.html) configuration for weekly or monthly schedules. See [refresh_on_day](#refresh_on_day).

### refresh_on_day

* `day_of_month` - (Optional) The day of the month that you want to schedule refresh on.
* `day_of_week` - (Optional) The day of the week that you want to schedule a refresh on. Valid values are `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY` and `SATURDAY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the refresh schedule.
* `id` - A comma-delimited string joining AWS account ID, data set ID & refresh schedule ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Refresh Schedule using the AWS account ID, data set ID and schedule ID separated by commas (`,`). For example:

```terraform
import {
  to = aws_quicksight_refresh_schedule.example
  id = "123456789012,dataset-id,schedule-id"
}
```

Using `terraform import`, import a QuickSight Refresh Schedule using the AWS account ID, data set ID and schedule ID separated by commas (`,`). For example:

```console
% terraform import aws_quicksight_refresh_schedule.example 123456789012,dataset-id,schedule-id
```
