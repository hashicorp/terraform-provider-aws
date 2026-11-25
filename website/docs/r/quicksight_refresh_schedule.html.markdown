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

* `data_set_id` - (Required, Forces new resource) ID of the dataset.
* `schedule` - (Required) [Refresh schedule](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RefreshSchedule.html) configuration. See [`schedule` Block](#schedule-block).
* `schedule_id` - (Required, Forces new resource) ID of the refresh schedule.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `schedule` Block

* `refresh_type` - (Required) Type of refresh that the dataset undergoes. Valid values are `INCREMENTAL_REFRESH` and `FULL_REFRESH`.
* `schedule_frequency` - (Optional) Configuration of the [schedule frequency](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RefreshFrequency.html). See [`schedule_frequency` Block](#schedule_frequency-block).
* `start_after_date_time` - (Optional) Time after which the refresh schedule can be started, expressed in `YYYY-MM-DDTHH:MM:SS` format.

### `schedule_frequency` Block

* `interval` - (Required) Interval between scheduled refreshes. Valid values are `MINUTE15`, `MINUTE30`, `HOURLY`, `DAILY`, `WEEKLY` and `MONTHLY`.
* `refresh_on_day` - (Optional) [Refresh on entity](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScheduleRefreshOnEntity.html) configuration for weekly or monthly schedules. See [`refresh_on_day` Block](#refresh_on_day-block).
* `time_of_the_day` - (Optional) Time of day that you want the dataset to refresh. This value is expressed in `HH:MM` format. This field is not required for schedules that refresh hourly.
* `timezone` - (Optional) Timezone that you want the refresh schedule to use.

### `refresh_on_day` Block

* `day_of_month` - (Optional) Day of the month that you want to schedule refresh on.
* `day_of_week` - (Optional) Day of the week that you want to schedule a refresh on. Valid values are `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY` and `SATURDAY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the refresh schedule.
* `id` - Comma-delimited string joining AWS account ID, data set ID & refresh schedule ID.

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
