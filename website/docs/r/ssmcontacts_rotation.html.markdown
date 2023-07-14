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
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"01:00"
	]
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
    number_of_on_calls = 1
	recurrence_multiplier = 1
	weekly_settings {
		day_of_week = "MON"
		hand_off_time = "04:25"
	}
	weekly_settings {
		day_of_week = "WED"
		hand_off_time = "07:34"
	}
	weekly_settings {
		day_of_week = "FRI"
		hand_off_time = "15:57"
	}
    shift_coverages {
		day_of_week = "MON"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
		}
  	}
	shift_coverages {
		day_of_week = "WED"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
		}
  	}
	shift_coverages {
		day_of_week = "FRI"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
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
    number_of_on_calls = 1
	recurrence_multiplier = 1
	monthly_settings {
		day_of_month = 20
		hand_off_time = "08:00"
	}
	monthly_settings {
		day_of_month = 13
		hand_off_time = "12:34"
	}
	monthly_settings {
		day_of_month = 1
		hand_off_time = "04:58"
	}
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}
```

## Argument Reference

~> **NOTE:** A rotation implicitly depends on a replication set. If you configured your replication set in Terraform, we recommend you add it to the `depends_on` argument for the Terraform Contact Resource.

The following arguments are required:

* `contact_ids` - (Required) Amazon Resource Names (ARNs) of the contacts to add to the rotation. The order in which you list the contacts is their shift order in the rotation schedule.
* `name` - (Required) The name for the rotation.
* `time_zone_id` - (Required) The time zone to base the rotation’s activity on in Internet Assigned Numbers Authority (IANA) format.
* `recurrence` - (Required) Information about when an on-call rotation is in effect and how long the rotation period lasts. Exactly one of either `daily_settings`, `monthly_settings`, or `weekly_settings` must be populated.
    * `number_of_oncalls` - (Required) The number of contacts, or shift team members designated to be on call concurrently during a shift.
    * `recurrence_multiplier` - (Required) The number of days, weeks, or months a single rotation lasts.
    * `daily_settings` - (Optional) Information about on-call rotations that recur daily. Composed of a list of times, in 24-hour format, for when daily on-call shift rotations begin.
    * `monthly_settings` - (Optional) Information about on-call rotations that recur monthly.
        * `day_of_month` - (Required) The day of the month when monthly recurring on-call rotations begin.
        * `hand_off_time` - (Required) The time of day, in 24-hour format, when a monthly recurring on-call shift rotation begins.
    * `weekly_settings` - (Optional) Information about rotations that recur weekly.
        * `day_of_week` - (Required) The day of the week when weekly recurring on-call shift rotations begins.
        * `hand_off_time` - (Required) The time of day, in 24-hour format, when a weekly recurring on-call shift rotation begins.
    * `shift_coverages` - (Optional) Information about the days of the week that the on-call rotation coverage includes.
        * `coverage_times` - (Required) Information about when an on-call shift begins and ends.
            * `end_time` - (Optional) The time, in 24-hour format, that the on-call rotation shift ends.
            * `start_time` - (Optional) The time, in 24-hour format, that the on-call rotation shift begins.
        * `day_of_week` - (Required) The day of the week when the shift coverage occurs.

The following arguments are optional:

* `start_time` - (Optional) The date and time, in RFC 3339 format, that the rotation goes into effect.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the rotation.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

An Incident Manager Contacts Rotation can be imported using the `ARN`. For example:

```
$ terraform import aws_ssmcontacts_rotation.example {ARNValue}
```
