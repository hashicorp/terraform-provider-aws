---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_rotation"
description: |-
  Provides a Terraform data source for managing a Contacts Rotation in AWS Systems Manager Incident Manager
---

# Data Source: aws_ssmcontacts_rotation

Provides a Terraform data source for managing a Contacts Rotation in AWS Systems Manager Incident Manager

## Example Usage

### Basic Usage

```terraform
data "aws_ssmcontacts_rotation" "example" {
  arn = "exampleARN"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) The Amazon Resource Name (ARN) of the rotation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `contact_ids` - The Amazon Resource Names (ARNs) of the contacts to add to the rotation. The order in which you list the contacts is their shift order in the rotation schedule.
* `name` - The name for the rotation.
* `time_zone_id` - The time zone to base the rotation’s activity on in Internet Assigned Numbers Authority (IANA) format.
* `recurrence` - Information about when an on-call rotation is in effect and how long the rotation period lasts. Exactly one of either `daily_settings`, `monthly_settings`, or `weekly_settings` must be populated.
    * `number_of_oncalls` - The number of contacts, or shift team members designated to be on call concurrently during a shift.
    * `recurrence_multiplier` - The number of days, weeks, or months a single rotation lasts.
    * `daily_settings` - - Information about on-call rotations that recur daily. Composed of a list of times, in 24-hour format, for when daily on-call shift rotation begins.
    * `monthly_settings` - Information about on-call rotations that recur monthly.
        * `day_of_month` - The day of the month when monthly recurring on-call rotations begin.
        * `hand_off_time`  The time of day, in 24-hour format, when a monthly recurring on-call shift rotation begins.
    * `weekly_settings` - Information about rotations that recur weekly.
        * `day_of_week` - The day of the week when weekly recurring on-call shift rotations begins.
        * `hand_off_time` - The time of day, in 24-hour format, when a weekly recurring on-call shift rotation begins.
    * `shift_coverages` - Information about the days of the week that the on-call rotation coverage includes.
        * `coverage_times` - Information about when an on-call shift begins and ends.
            * `end_time` - The time, in 24-hour format,, hat the on-call rotation shift ends.
            * `start_time` - The time, in 24-hour format, that the on-call rotation shift begins.
        * `day_of_week` - The day of the week when the shift coverage occurs.
* `start_time` - The date and time, in RFC 3339 format, that the rotation goes into effect.
* `tags` - A map of tags to assign to the resource.
