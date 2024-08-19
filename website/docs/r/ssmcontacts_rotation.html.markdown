---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_rotation"
description: |-
  Provides a Terraform resource for managing a Contacts Rotation in AWS Systems Manager Incident Manager.
---

# Resource: aws_ssmcontacts_rotation

Provides a Terraform resource for managing a Contacts Rotation in AWS Systems Manager Incident Manager.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmcontacts_rotation" "example" {
  contact_ids = [
    aws_ssmcontacts_contact.example.arn
  ]

  name = "rotation"

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 00
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

### Usage with Weekly Settings and Shift Coverages Fields

```terraform
resource "aws_ssmcontacts_rotation" "example" {
  contact_ids = [
    aws_ssmcontacts_contact.example.arn
  ]

  name = "rotation"

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    weekly_settings {
      day_of_week = "WED"
      hand_off_time {
        hour_of_day    = 04
        minute_of_hour = 25
      }
    }

    weekly_settings {
      day_of_week = "FRI"
      hand_off_time {
        hour_of_day    = 15
        minute_of_hour = 57
      }
    }

    shift_coverages {
      map_block_key = "MON"
      coverage_times {
        start {
          hour_of_day    = 01
          minute_of_hour = 00
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 00
        }
      }
    }
  }

  start_time = "2023-07-20T02:21:49+00:00"

  time_zone_id = "Australia/Sydney"

  tags = {
    key1 = "tag1"
    key2 = "tag2"
  }

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

### Usage with Monthly Settings Fields

```terraform
resource "aws_ssmcontacts_rotation" "example" {
  contact_ids = [
    aws_ssmcontacts_contact.example.arn,
  ]

  name = "rotation"

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    monthly_settings {
      day_of_month = 20
      hand_off_time {
        hour_of_day    = 8
        minute_of_hour = 00
      }
    }
    monthly_settings {
      day_of_month = 13
      hand_off_time {
        hour_of_day    = 12
        minute_of_hour = 34
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.example]
}
```

## Argument Reference

~> **NOTE:** A rotation implicitly depends on a replication set. If you configured your replication set in Terraform, we recommend you add it to the `depends_on` argument for the Terraform Contact Resource.

The following arguments are required:

* `contact_ids` - (Required) Amazon Resource Names (ARNs) of the contacts to add to the rotation. The order in which you list the contacts is their shift order in the rotation schedule.
* `name` - (Required) The name for the rotation.
* `time_zone_id` - (Required) The time zone to base the rotation’s activity on in Internet Assigned Numbers Authority (IANA) format.
* `recurrence` - (Required) Information about when an on-call rotation is in effect and how long the rotation period lasts. Exactly one of either `daily_settings`, `monthly_settings`, or `weekly_settings` must be populated. See [Recurrence](#recurrence) for more details.

The following arguments are optional:

* `start_time` - (Optional) The date and time, in RFC 3339 format, that the rotation goes into effect.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the rotation.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### Recurrence

* `number_of_on_calls` - (Required) The number of contacts, or shift team members designated to be on call concurrently during a shift.
* `recurrence_multiplier` - (Required) The number of days, weeks, or months a single rotation lasts.
* `daly_settings` - (Optional) Information about on-call rotations that recur daily. Composed of a list of times, in 24-hour format, for when daily on-call shift rotations begin. See [Daily Settings](#daily-settings) for more details.
* `monthly_settings` - (Optional) Information about on-call rotations that recur monthly. See [Monthly Settings](#monthly-settings) for more details.
* `weekly_settings` - (Optional) Information about on-call rotations that recur weekly. See [Weekly Settings](#weekly-settings) for more details.
* `shift_coverages` - (Optional) Information about the days of the week that the on-call rotation coverage includes. See [Shift Coverages](#shift-coverages) for more details.

### Daily Settings

* `hour_of_day` - (Required) The hour of the day.
* `minute_of_hour` - (Required) The minutes of the hour.

### Monthly Settings

* `day_of_month` - (Required) The day of the month when monthly recurring on-call rotations begin.
* `hand_off_time` - (Required) The hand off time. See [Hand Off Time](#hand-off-time) for more details.

### Weekly Settings

* `day_of_week` - (Required) The day of the week when weekly recurring on-call shift rotations begins.
* `hand_off_time` - (Required) The hand off time. See [Hand Off Time](#hand-off-time) for more details.

### Hand Off Time

* `hour_of_day` - (Required) The hour of the day.
* `minute_of_hour` - (Required) The minutes of the hour.

### Shift Coverages

* `coverage_times` - (Required) Information about when an on-call shift begins and ends. See [Coverage Times](#coverage-times) for more details.
* `day_of_week` - (Required) The day of the week when the shift coverage occurs.

### Coverage Times

* `start` - (Required) The start time of the on-call shift. See [Hand Off Time](#hand-off-time) for more details.
* `end` - (Required) The end time of the on-call shift. See [Hand Off Time](#hand-off-time) for more details.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSMContacts Rotation using the `arn`. For example:

```terraform
import {
  to = aws_ssmcontacts_rotation.example
  id = "arn:aws:ssm-contacts:us-east-1:012345678910:rotation/example"
}
```

Using `terraform import`, import CodeGuru Profiler Profiling Group using the `arn`. For example:

```console
% terraform import aws_ssmcontacts_rotation.example arn:aws:ssm-contacts:us-east-1:012345678910:rotation/example
```
