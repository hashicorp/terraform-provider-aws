---
subcategory: "Application Signals"
layout: "aws"
page_title: "AWS: aws_applicationsignals_service_level_objective"
description: |-
  Provides details about an AWS Application Signals Service Level Objective.
---
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Data Source: aws_applicationsignals_service_level_objective

Provides details about an AWS Application Signals Service Level Objective.

## Example Usage

### Basic Usage
### Using Service Level Objective Name
```terraform

data "aws_applicationsignals_service_level_objective" "example" {
  id = "example-slo"
}
```

### Using Service Level Objective ARN
```terraform

data "aws_applicationsignals_service_level_objective" "example" {
  id = "arn:aws:application-signals:REGION:ACCOUNT_ID:slo/example-slo"
}
```
## Argument Reference

The following arguments are required:

* `id` - (Required) Accepts either the ARN or name of the Service Level Objective.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Level Objective.
* `created_time` - Timestamp when the SLO was created (RFC3339 format).
* `last_updated_time` - Timestamp of the last update to the SLO (RFC3339 format).
* `description` - Description of the SLO.
* `name` - Name of the SLO.
* `metric_source_type` - Type of metric source used for the SLO.
* `evaluation_type` - Type of evaluation used for the SLO.
* `id` - ID of the SLO.
* `goal` - Nested block defining the goal configuration:
    * `attainment_goal` - Target goal value for the SLO.
    * `warning_threshold` - Warning threshold for the SLO.
    * `interval` - Interval configuration:
        * `calendar_interval`:
            * `duration` - Duration of the calendar interval.
            * `duration_unit` - Unit of duration (e.g., seconds, minutes).
            * `start_time` - Start time of the interval.
        * `rolling_interval`:
            * `duration` - Duration of the rolling interval.
            * `duration_unit` - Unit of duration.
* `burn_rate_configurations` - List of burn rate configurations:
    * `look_back_window_minutes` - Look-back window in minutes for burn rate calculation.
* `request_based_sli` - Request-based SLI configuration:
    * `metric_threshold` - Metric threshold for request-based SLI.
    * `comparison_operator` - Comparison operator used for evaluation.
    * `request_based_sli_metric`:
        * `key_attributes` - Map of key attributes for the metric.
        * `metric_type` - Type of the metric.
        * `operation_name` - Name of the operation.
        * `dependency_config` - Dependency configuration for the metric.
        * `total_request_count_metric` - List of metric data queries (see `metric_data_queries` below).
* `sli` - SLI configuration:
    * `metric_threshold` - Metric threshold for the SLI.
    * `comparison_operator` - Comparison operator for the SLI.
    * `sli_metric`:
        * `key_attributes` - Map of key attributes for the metric.
        * `metric_type` - Type of the metric.
        * `operation_name` - Name of the operation.
        * `dependency_config` - Dependency configuration for the metric.
        * `metric_data_queries` - List of metric data queries:
            * `id` - ID of the metric data query.
            * `account_id` - AWS account ID for the metric.
            * `expression` - Metric math expression.
            * `label` - Label for the metric query.
            * `period` - Period for the metric in seconds.
            * `return_data` - Boolean indicating if data is returned.
            * `metric_stat`:
                * `period` - Period for the metric stat.
                * `stat` - Statistic type (e.g., Average, Sum).
                * `unit` - Unit of the metric.
                * `metric`:
                    * `metric_name` - Name of the metric.
                    * `namespace` - Namespace of the metric.
                    * `dimensions` - List of dimensions:
                        * `name` - Name of the dimension.
                        * `value` - Value of the dimension.
