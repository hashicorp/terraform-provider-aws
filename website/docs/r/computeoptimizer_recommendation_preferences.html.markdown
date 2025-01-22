---
subcategory: "Compute Optimizer"
layout: "aws"
page_title: "AWS: aws_computeoptimizer_recommendation_preferences"
description: |-
  Manages AWS Compute Optimizer recommendation preferences.
---

# Resource: aws_computeoptimizer_recommendation_preferences

Manages AWS Compute Optimizer recommendation preferences.

## Example Usage

### Lookback Period Preference

```terraform
resource "aws_computeoptimizer_recommendation_preferences" "example" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = "123456789012"
  }

  look_back_period = "DAYS_32"
}
```

### Multiple Preferences

```terraform
resource "aws_computeoptimizer_recommendation_preferences" "example" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = "123456789012"
  }

  enhanced_infrastructure_metrics = "Active"

  external_metrics_preference {
    source = "Datadog"
  }

  preferred_resource {
    include_list = ["m5.xlarge", "r5"]
    name         = "Ec2InstanceTypes"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `enhanced_infrastructure_metrics` - (Optional) The status of the enhanced infrastructure metrics recommendation preference. Valid values: `Active`, `Inactive`.
* `external_metrics_preference` - (Optional) The provider of the external metrics recommendation preference. See [External Metrics Preference](#external-metrics-preference) below.
* `inferred_workload_types` - (Optional) The status of the inferred workload types recommendation preference. Valid values: `Active`, `Inactive`.
* `look_back_period` - (Optional) The preference to control the number of days the utilization metrics of the AWS resource are analyzed. Valid values: `DAYS_14`, `DAYS_32`, `DAYS_93`.
* `preferred_resource` - (Optional) The preference to control which resource type values are considered when generating rightsizing recommendations. See [Preferred Resources](#preferred-resources) below.
* `resource_type` - (Required) The target resource type of the recommendation preferences. Valid values: `Ec2Instance`, `AutoScalingGroup`, `RdsDBInstance`.
* `savings_estimation_mode` - (Optional) The status of the savings estimation mode preference. Valid values: `AfterDiscounts`, `BeforeDiscounts`.
* `scope` - (Required) The scope of the recommendation preferences. See [Scope](#scope) below.
* `utilization_preference` - (Optional) The preference to control the resource’s CPU utilization threshold, CPU utilization headroom, and memory utilization headroom. See [Utilization Preferences](#utilization-preferences) below.

### External Metrics Preference

* `source` - (Required) The source options for external metrics preferences. Valid values: `Datadog`, `Dynatrace`, `NewRelic`, `Instana`.

### Preferred Resources

You can specify this preference as a combination of include and exclude lists.
You must specify either an `include_list` or `exclude_list`.

* `exclude_list` - (Optional) The preferred resource type values to exclude from the recommendation candidates. If this isn’t specified, all supported resources are included by default.
* `include_list` - (Optional) The preferred resource type values to include in the recommendation candidates. You can specify the exact resource type value, such as `"m5.large"`, or use wild card expressions, such as `"m5"`. If this isn’t specified, all supported resources are included by default.
* `name` - (Required) The type of preferred resource to customize. Valid values: `Ec2InstanceTypes`.

### Scope

* `name` - (Required) The name of the scope. Valid values: `Organization`, `AccountId`, `ResourceArn`.
* `value` - (Required) The value of the scope. `ALL_ACCOUNTS` for `Organization` scopes, AWS account ID for `AccountId` scopes, ARN of an EC2 instance or an Auto Scaling group for `ResourceArn` scopes.

### Utilization Preferences

* `metric_name` - (Required) The name of the resource utilization metric name to customize. Valid values: `CpuUtilization`, `MemoryUtilization`.
* `metric_parameters` - (Required) The parameters to set when customizing the resource utilization thresholds.
    * `headroom` - (Required) The headroom value in percentage used for the specified metric parameter. Valid values: `PERCENT_30`, `PERCENT_20`, `PERCENT_10`, `PERCENT_0`.
    * `threshold` - (Optional) The threshold value used for the specified metric parameter. You can only specify the threshold value for CPU utilization. Valid values: `P90`, `P95`, `P99_5`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import recommendation preferences using the resource type, scope name and scope value. For example:

```terraform
import {
  to = aws_computeoptimizer_recommendation_preferences.example
  id = "Ec2Instance,AccountId,123456789012"
}
```

Using `terraform import`, import recommendation preferences using the resource type, scope name and scope value. For example:

```console
% terraform import aws_computeoptimizer_recommendation_preferences.example Ec2Instance,AccountId,123456789012
```
