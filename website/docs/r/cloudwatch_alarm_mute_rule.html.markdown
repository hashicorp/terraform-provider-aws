---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_alarm_mute_rule"
description: |-
  Manages an AWS CloudWatch Alarm Mute Rule.
---

# Resource: aws_cloudwatch_alarm_mute_rule

Manages an AWS CloudWatch Alarm Mute Rule.

## Example Usage

### Basic Usage with Cron Expression

```terraform
resource "aws_cloudwatch_alarm_mute_rule" "example" {
  name = "example"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }
}
```

### With Start/Expire Dates Option

```terraform
resource "aws_cloudwatch_alarm" "example" {
  alarm_name          = "example"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
}

resource "aws_cloudwatch_alarm_mute_rule" "example" {
  name        = "example"
  description = "Mute alarms during maintenance window"
  start_date  = "2026-01-01T00:00:00Z"
  expire_date = "2026-12-31T23:59:00Z"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
      timezone   = "Asia/Tokyo"
    }
  }

  mute_targets {
    alarm_names = [aws_cloudwatch_alarm.example.alarm_name]
  }

  tags = {
    Environment = "production"
  }
}
```

### With At Expression

~> **NOTE:** When using `at()` expressions, do not set `start_date` or `expire_date`. The CloudWatch API returns the error `Can not set start or expire dates for At expressions`.

```terraform
resource "aws_cloudwatch_alarm_mute_rule" "example" {
  name = "example"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "at(2026-12-31T23:59:59)"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the alarm mute rule. Changing this forces a new resource.
* `rule` - (Required) Rule definition for the mute rule. See [`rule` block](#rule) below for details.

The following arguments are optional:

* `description` - (Optional) Description of the alarm mute rule.
* `expire_date` - (Optional) Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) when the mute rule expires. Seconds must be set to `00` (e.g., `2026-12-31T23:59:00Z`). Must not be set when using `at()` expressions.
* `mute_targets` - (Optional) Alarms to mute. See [`mute_targets` block](#mute_targets) below for details.
* `start_date` - (Optional) Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) when the mute rule becomes active. Seconds must be set to `00` (e.g., `2026-01-01T00:00:00Z`). Must not be set when using `at()` expressions.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rule`

* `schedule` - (Required) Schedule for the mute rule. See [`schedule` block](#schedule) below for details.

### `schedule`

* `duration` - (Required) Duration of the mute period in [ISO 8601 duration format](https://en.wikipedia.org/wiki/ISO_8601#Durations) (e.g., `PT4H` for 4 hours).
* `expression` - (Required) Schedule expression. Supports `cron()` and `at()` formats. For example, `cron(0 2 * * *)` for daily at 2:00 AM or `at(2026-01-01T00:00)` for a one-time mute. See [Defining alarm mute rules](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/alarm-mute-rules.html#defining-alarm-mute-rules) for details.
* `timezone` - (Optional) Timezone for the schedule expression (e.g., `Asia/Tokyo`). Defaults to UTC.

### `mute_targets`

* `alarm_names` - (Required) List of alarm names to mute.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Alarm Mute Rule.
* `last_updated_timestamp` - Timestamp of when the mute rule was last updated.
* `mute_type` - Indicates whether the mute rule is one-time or recurring. Valid values are `ONE_TIME` or `RECURRING`. See [Alarm mute rules](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/alarm-mute-rules.html) for details.
* `status` - Current status of the mute rule. Valid values are `SCHEDULED`, `ACTIVE`, or `EXPIRED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_alarm_mute_rule.example
  identity = {
    name = "example"
  }
}

resource "aws_cloudwatch_alarm_mute_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the CloudWatch Alarm Mute Rule.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Alarm Mute Rule using the `name`. For example:

```terraform
import {
  to = aws_cloudwatch_alarm_mute_rule.example
  id = "example"
}
```

Using `terraform import`, import CloudWatch Alarm Mute Rule using the `name`. For example:

```console
% terraform import aws_cloudwatch_alarm_mute_rule.example example
```
