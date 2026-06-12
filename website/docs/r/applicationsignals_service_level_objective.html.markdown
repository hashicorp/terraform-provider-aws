---
subcategory: "Application Signals"
layout: "aws"
page_title: "AWS: aws_applicationsignals_service_level_objective"
description: |-
  Manages an AWS Application Signals Service Level Objective.
---

# Resource: aws_applicationsignals_service_level_objective

Manages an AWS Application Signals Service Level Objective.

## Example Usage

### Basic Usage with a Period-Based SLO

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
  name        = "elb-error-rate"
  description = "Error rate of 99.98% for 90 days"
  goal {
    interval {
      rolling_interval {
        duration_unit = "DAY"
        duration      = 90
      }
    }
    attainment_goal   = 99.98
    warning_threshold = 30.0
  }
  sli {
    comparison_operator = "LessThan"
    metric_threshold    = 2
    sli_metric {
      metric_data_queries {
        id = "m1"
        metric_stat {
          metric {
            namespace   = "AWS/ApplicationELB"
            metric_name = "HTTPCode_Target_5XX_Count"
            dimensions {
              name  = "LoadBalancer"
              value = "app/my-load-balancer"
            }
          }
          period = 300
          stat   = "Sum"
        }
        return_data = true
      }
    }
  }
}
```

### Request-Based SLO Usage

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
  name        = "lambda-success-rate"
  description = "Success rate of 99.9% for a specific operation over a calendar month"
  goal {
    interval {
      rolling_interval {
        duration      = 1
        duration_unit = "DAY"
      }
    }
    attainment_goal   = 99.90
    warning_threshold = 50.0
  }
  request_based_sli {
    request_based_sli_metric {
      total_request_count_metric {
        metric_stat {
          metric {
            namespace  = "AWS/Lambda"
            metric_name = "Invocations"
            dimensions {
              name  = "Dimension1"
              value = "my-dimension-name"
            }
          }
          period = 60
          stat = "Sum"
        }
        id = "total_requests"
        return_data = true
      }
      monitored_request_count_metric {
        bad_count_metric {
          metric_stat {
            metric {
              namespace   = "AWS/Lambda"
              metric_name = "ErrorCount"
            }
            period = 60
            stat   = "Sum"
          }
          id = "bad_requests"
          return_data = true
        }
      }
    }
  }
}
```

-----

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of this SLO. Must be unique for your AWS account and is immutable after creation.
* [`goal`](#goal) - (Required) Configuration block determining the goal of this SLO.

The following arguments are optional:

* `description` - (Optional) Brief description of the SLO.
* [`burn_rate_configurations`](#burn_rate_configurations) - (Optional) Configuration block containing attributes that determine the burn rates of this SLO.
* [`request_based_sli`](#request_based_sli) - (Optional) Configuration block for a request-based Service Level Indicator (SLI).
* [`sli`](#sli) - (Optional) Configuration block for a period-based Service Level Indicator (SLI).
* `timeouts` - (Optional) Configuration block for setting operation timeouts.

> You must specify exactly one `sli` or `request_based_sli`.

## Block Reference

### burn_rate_configurations

* `look_back_window_minutes` - (Required) The number of minutes to use as the look back window for calculating the burn rate.

### goal

* `attainment_goal` - (Required) The threshold that determines if the goal is being met.
* [`interval`](#interval) - (Required) Configuration block defining the time period used to evaluate the SLO.
* `warning_threshold` - (Required) The percentage of remaining budget over total budget that you want to get warnings for.

### interval

The `interval` block must contain exactly one of the following blocks:

* [`calendar_interval`](#calendar_interval) - Configuration block for a time interval that starts at a specific time and runs for a specified duration.
* [`rolling_interval`](#rolling_interval) - Configuration block for a time interval that rolls forward by a specified duration.

### calendar_interval

* `duration` - (Required) The duration of the calendar interval.
* `duration_unit` - (Required) The unit of time for the duration (`MINUTE`, `HOUR`, `DAY`, `MONTH`).
* `start_time` - (Required) The date and time when you want the first interval to start in **RFC3339** format (e.g., `2024-01-01T00:00:00Z`).

### rolling_interval

* `duration` - (Required) The duration of the rolling interval.
* `duration_unit` - (Required) The unit of time for the duration (`MINUTE`, `HOUR`, `DAY`, `MONTH`).

### sli

* `comparison_operator` - (Optional) The arithmetic operation to use when comparing the specified metric to the threshold.
* `metric_threshold` - (Optional) The value the SLI metric value is compared to.
* [`sli_metric`](#sli_metric) - (Optional) Configuration block defining the metric for this period-based SLI.

### sli_metric

* [`dependency_config`](#dependency_config) - (Optional) Configuration block for identifying the dependency.
* `key_attributes` - (Optional) A map of key-value pairs to specify which service this SLO metric is related to.
* [`metric_data_queries`](#metric_data_queries) - (Optional) Configuration block for a list of CloudWatch metric data queries.
* `metric_name` - (Optional) The name of the CloudWatch metric to use.
* `metric_type` - (Optional) The metric type that Application Signals collects. Must be either `AVAILABILITY` or `LATENCY`.
* `operation_name` - (Optional) If the SLO is to monitor a specific operation of the service, use this field to specify the name of that operation.
* `period_seconds` - (Optional) The number of seconds to use as the period for the CloudWatch metric.
* `statistic` - (Optional) The statistic to use for comparison to the threshold.

### request_based_sli

* `comparison_operator` - (Optional) The arithmetic operation to use when comparing the specified metric to the threshold.
* `metric_threshold` - (Optional) The percentage success rate the comparison operator is compared to.
* [`request_based_sli_metric`](#request_based_sli_metric) - (Optional) Configuration block defining the metrics for this request-based SLI.

### request_based_sli_metric

* [`dependency_config`](#dependency_config) - (Optional) Configuration block for identifying the dependency.
* `key_attributes` - (Optional) A map of key-value pairs to specify which service this SLO metric is related to.
* `metric_type` - (Optional) The metric type that Application Signals collects. Must be either `AVAILABILITY` or `LATENCY`.
* [`monitored_request_count_metric`](#monitored_request_count_metric) - (Optional) Configuration block defining the good or bad request value for a request-based SLO.
* `operation_name` - (Optional) If the SLO is to monitor a specific operation of the service, use this field to specify the name of that operation.
* [`total_request_count_metric`](#total_request_count_metric) - (Optional) Configuration block for the metric to be used as the total requests for a request-based SLO.

### monitored_request_count_metric

* [`good_count_metric`](#good_count_metric) - (Optional) Configuration block for the metric that counts good requests.
* [`bad_count_metric`](#bad_count_metric) - (Optional) Configuration block for the metric that counts bad requests.

### good_count_metric

You must specify either `expression` or `metric_stat` but not both.

* `account_id` - (Optional) The ID of the account where this metric is located.
* `expression` - (Optional) A metric math expression to be performed on the other metrics.
* `id` - (Optional) An ID (unique within the outer block) for the metric data query.
* `label` - (Optional) A human-readable label for this metric or expression.
* `period` - (Optional) The granularity, in seconds, of the returned data points for this metric.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a metric to be used directly for the SLO, or to be used in the math expression that will be used for the SLO.

### bad_count_metric

You must specify either `expression` or `metric_stat` but not both.

* `account_id` - (Optional) The ID of the account where this metric is located.
* `expression` - (Optional) A metric math expression to be performed on the other metrics.
* `id` - (Optional) An ID (unique within the outer block) for the metric data query.
* `label` - (Optional) A human-readable label for this metric or expression.
* `period` - (Optional) The granularity, in seconds, of the returned data points for this metric.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a metric to be used directly for the SLO, or to be used in the math expression that will be used for the SLO.

### total_request_count_metric

You must specify either `expression` or `metric_stat` but not both.

* `account_id` - (Optional) The ID of the account where this metric is located.
* `expression` - (Optional) A metric math expression to be performed on the other metrics.
* `id` - (Optional) An ID (unique within the outer block) for the metric data query.
* `label` - (Optional) A human-readable label for this metric or expression.
* `period` - (Optional) The granularity, in seconds, of the returned data points for this metric.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a metric to be used directly for the SLO, or to be used in the math expression that will be used for the SLO.

### dependency_config

* `dependency_key_attributes` - (Required) A map of key-value pairs to identify the dependency.
* `dependency_operation_name` - (Required) The name of the called operation in the dependency.

### metric_data_queries

You must specify either `expression` or `metric_stat` but not both.

* `account_id` - (Optional) The ID of the account where this metric is located.
* `expression` - (Optional) A metric math expression to be performed on the other metrics.
* `id` - (Optional) An ID (unique within the outer block) for the metric data query.
* `label` - (Optional) A human-readable label for this metric or expression.
* `period` - (Optional) The granularity, in seconds, of the returned data points for this metric.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a metric to be used directly for the SLO, or to be used in the math expression that will be used for the SLO.

### metric_stat

* [`metric`](#metric) - (Optional) Configuration block for the metric.
* `period` - (Optional) The period over which the metric is aggregated.
* `stat` - (Optional) The statistic to apply to the metric.
* `unit` - (Optional) The unit for the metric.

### metric

* [`dimensions`](#dimensions) - (Optional) A configuration block defining one or more dimensions to use to define the metric that you want to use.
* `metric_name` - (Optional) The name of the metric to use.
* `namespace` - (Optional) The namespace of the metric.

### dimensions

* `name` - (Required) The name of the dimension.
* `value` - (Required) The value of the dimension.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Level Objective.
* `created_time` - The date and time that this SLO was created (RFC3339 format).
* `last_updated_time` - The time that this SLO was most recently updated (RFC3339 format).
* `evaluation_type` - Displays whether this is a period-based SLO or a request-based SLO.
* `metric_source_type` - Displays the source of the SLI metric for this SLO.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Signals Service Level Objective using its `name`. For example:

```terraform
import {
  to = aws_applicationsignals_service_level_objective.example
  id = "my-slo-name"
}
```

Using `terraform import`, import Application Signals Service Level Objective using the `name`. For example:

```console
% terraform import aws_applicationsignals_service_level_objective.example my-slo-name
```
