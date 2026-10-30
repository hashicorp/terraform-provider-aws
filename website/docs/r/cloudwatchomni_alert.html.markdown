---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_alert"
description: |-
  Manages an Amazon CloudWatch Omni alert.
---

# Resource: aws_cloudwatchomni_alert

Manages an Amazon CloudWatch Omni alert.

An alert periodically evaluates a telemetry query in a space and sends notifications when its state changes.

## Example Usage

### Basic Usage

An alert's access profile must have an access grant that lets alerts assume it.

```terraform
resource "aws_cloudwatchomni_access_grant" "example" {
  name       = "alerts-assume-example"
  domain_id  = aws_cloudwatchomni_domain.example.domain_id
  space_id   = aws_cloudwatchomni_space.example.space_id
  permission = "CUSTOM"

  principal {
    principal_type = "ALERT"
    principal_id   = "ALL"
  }

  scoped_action {
    actions = ["cloudwatch:AssumeAccessProfile"]

    resource {
      resource_type = "AccessProfile"
      resource_arns = [aws_cloudwatchomni_access_profile.example.arn]
    }
  }
}

resource "aws_cloudwatchomni_alert" "example" {
  space_id   = aws_cloudwatchomni_space.example.space_id
  profile_id = aws_cloudwatchomni_access_profile.example.profile_id
  name       = "example"

  rule {
    telemetry_rule {
      query {
        expression = "up"
        language   = "PROMQL"
      }

      condition {
        comparator         = "GT"
        threshold_mode     = "COUNT_OF_RESULTS"
        critical_threshold = 0
      }

      evaluation {
        interval_seconds = 60
      }

      no_data {
        treat_as = "OK"
      }
    }
  }

  depends_on = [aws_cloudwatchomni_access_grant.example]
}
```

### With Notifications

The SNS topic's access policy must allow the `cloudwatch.amazonaws.com` service principal to publish to it. The access grant is the same as in the basic example.

```terraform
resource "aws_cloudwatchomni_alert" "example" {
  space_id    = aws_cloudwatchomni_space.example.space_id
  profile_id  = aws_cloudwatchomni_access_profile.example.profile_id
  name        = "example"
  description = "Fires when any target is down."

  rule {
    telemetry_rule {
      query {
        expression = "up == 0"
        language   = "PROMQL"
      }

      condition {
        comparator         = "GT"
        threshold_mode     = "COUNT_OF_RESULTS"
        critical_threshold = 0
      }

      evaluation {
        interval_seconds          = 60
        pending_duration_seconds  = 120
        recovery_duration_seconds = 300
      }

      no_data {
        treat_as = "OK"
      }
    }
  }

  notification_rule {
    target {
      arn  = aws_sns_topic.example.arn
      type = "sns"
    }

    trigger {
      state_values = ["CRITICAL", "OK"]
    }
  }

  depends_on = [aws_cloudwatchomni_access_grant.example]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the alert. Must be 1-256 characters. Names don't have to be unique within a space.
* `profile_id` - (Required, Forces new resource) ID of the access profile the alert uses to evaluate its query and send notifications. The access profile must have an access grant that lets alerts assume it: principal type `ALERT` with the custom action `cloudwatch:AssumeAccessProfile` (see `aws_cloudwatchomni_access_grant`).
* `rule` - (Required) Rule that defines how the alert is evaluated. [See below](#rule-block).
* `space_id` - (Required, Forces new resource) ID of the space that contains the alert.

The following arguments are optional:

* `description` - (Optional) Description of the alert. Removing the description replaces the alert, because the service does not support clearing it.
* `notification_rule` - (Optional) One or more rules that determine when and where notifications are sent. [See below](#notification_rule-block).
* `notifications_enabled` - (Optional) Whether notifications are enabled for the alert. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rule` Block

The `rule` configuration block supports the following arguments:

* `telemetry_rule` - (Required) Rule that evaluates a telemetry query. [See below](#telemetry_rule-block).

### `telemetry_rule` Block

The `telemetry_rule` configuration block supports the following arguments:

* `condition` - (Required) Condition that determines when the alert fires. [See below](#condition-block).
* `evaluation` - (Required) Evaluation cadence and durations. [See below](#evaluation-block).
* `no_data` - (Required) How the alert behaves when the query returns no data. [See below](#no_data-block).
* `query` - (Required) Query to evaluate. [See below](#query-block).

### `condition` Block

The `condition` configuration block supports the following arguments:

* `comparator` - (Optional) Comparison operator applied to the thresholds. Valid values: `GT`, `LT`, `GTE`, `LTE`.
* `critical_threshold` - (Optional) Value at which the alert enters the `CRITICAL` state.
* `threshold_field` - (Optional) Field of the query results that the thresholds are compared with.
* `threshold_mode` - (Optional) How the thresholds are applied to query results. Valid values: `COUNT_OF_RESULTS`, `FIELD_VALUE`.
* `warning_threshold` - (Optional) Value at which the alert enters the `WARNING` state.

### `evaluation` Block

The `evaluation` configuration block supports the following arguments:

* `interval_seconds` - (Required) Interval between evaluations, in seconds.
* `pending_duration_seconds` - (Optional) How long a breach must last before the alert fires, in seconds. Must be a multiple of `interval_seconds`.
* `recovery_duration_seconds` - (Optional) How long a recovery must last before the alert clears, in seconds. Must be a multiple of `interval_seconds`.

### `no_data` Block

The `no_data` configuration block supports the following arguments:

* `treat_as` - (Required) State to report when an evaluation returns no data. Valid values: `OK`, `WARNING`, `CRITICAL`, `NODATA`.

### `query` Block

The `query` configuration block supports the following arguments:

* `expression` - (Required) Query expression.
* `language` - (Required) Query language of the expression. Valid values: `SQL`, `PROMQL`.

### `notification_rule` Block

The `notification_rule` configuration block supports the following arguments:

* `target` - (Required) Destination for the notifications. [See below](#target-block).
* `trigger` - (Required) Alert states that trigger the notifications. [See below](#trigger-block).

### `target` Block

The `target` configuration block supports the following arguments:

* `arn` - (Required) ARN of the notification target. For `sns`, the ARN of an SNS topic. For `slack` and `pagerduty`, the ARN of a CloudWatch Omni integration.
* `metadata` - (Optional) Map of additional target-specific metadata. Must contain at least one entry if set.
* `type` - (Required) Type of the notification target. Valid values: `sns`, `slack`, `pagerduty`.

### `trigger` Block

The `trigger` configuration block supports the following arguments:

* `state_values` - (Optional) Set of alert states that trigger the notifications. Valid values: `OK`, `WARNING`, `CRITICAL`, `NODATA`. Omit to trigger on any state.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alert_id` - Unique identifier of the alert.
* `arn` - ARN of the alert.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_alert.example
  identity = {
    space_id = "12345678-1234-1234-1234-123456789012"
    alert_id = "0123456789abcdef0123456789abcdef"
  }
}

resource "aws_cloudwatchomni_alert" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `alert_id` (String) Alert ID.
* `space_id` (String) Space ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Alert using the `space_id` and `alert_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudwatchomni_alert.example
  id = "12345678-1234-1234-1234-123456789012,0123456789abcdef0123456789abcdef"
}
```

Using `terraform import`, import a CloudWatch Omni Alert using the `space_id` and `alert_id` separated by a comma (`,`). For example:

```console
% terraform import aws_cloudwatchomni_alert.example 12345678-1234-1234-1234-123456789012,0123456789abcdef0123456789abcdef
```
