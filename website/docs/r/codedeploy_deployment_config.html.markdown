---
subcategory: "CodeDeploy"
layout: "aws"
page_title: "AWS: aws_codedeploy_deployment_config"
description: |-
  Provides a CodeDeploy deployment config.
---

# Resource: aws_codedeploy_deployment_config

Provides a CodeDeploy deployment config for an application

## Example Usage

### Server Usage

```terraform
resource "aws_codedeploy_deployment_config" "foo" {
  deployment_config_name = "test-deployment-config"

  minimum_healthy_hosts {
    type  = "HOST_COUNT"
    value = 2
  }
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name               = aws_codedeploy_app.foo_app.name
  deployment_group_name  = "bar"
  service_role_arn       = aws_iam_role.foo_role.arn
  deployment_config_name = aws_codedeploy_deployment_config.foo.id

  ec2_tag_filter {
    key   = "filterkey"
    type  = "KEY_AND_VALUE"
    value = "filtervalue"
  }

  trigger_configuration {
    trigger_events     = ["DeploymentFailure"]
    trigger_name       = "foo-trigger"
    trigger_target_arn = "foo-topic-arn"
  }

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  alarm_configuration {
    alarms  = ["my-alarm-name"]
    enabled = true
  }
}
```

### Lambda Usage

```terraform
resource "aws_codedeploy_deployment_config" "foo" {
  deployment_config_name = "test-deployment-config"
  compute_platform       = "Lambda"

  traffic_routing_config {
    type = "TimeBasedLinear"

    time_based_linear {
      interval   = 10
      percentage = 10
    }
  }
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name               = aws_codedeploy_app.foo_app.name
  deployment_group_name  = "bar"
  service_role_arn       = aws_iam_role.foo_role.arn
  deployment_config_name = aws_codedeploy_deployment_config.foo.id

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_STOP_ON_ALARM"]
  }

  alarm_configuration {
    alarms  = ["my-alarm-name"]
    enabled = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `deployment_config_name` - (Required) The name of the deployment config.
* `compute_platform` - (Optional) The compute platform can be `Server`, `Lambda`, or `ECS`. Default is `Server`.
* `minimum_healthy_hosts` - (Optional) A minimum_healthy_hosts block. Required for `Server` compute platform. Minimum Healthy Hosts are documented below.
* `traffic_routing_config` - (Optional) A traffic_routing_config block. Traffic Routing Config is documented below.
* `zonal_config` - (Optional) A zonal_config block. Zonal Config is documented below.

The `minimum_healthy_hosts` block supports the following:

* `type` - (Required) The type can either be `FLEET_PERCENT` or `HOST_COUNT`.
* `value` - (Required) The value when the type is `FLEET_PERCENT` represents the minimum number of healthy instances as
a percentage of the total number of instances in the deployment. If you specify FLEET_PERCENT, at the start of the
deployment, AWS CodeDeploy converts the percentage to the equivalent number of instance and rounds up fractional instances.
When the type is `HOST_COUNT`, the value represents the minimum number of healthy instances as an absolute value.

The `traffic_routing_config` block supports the following:

* `type` - (Optional) Type of traffic routing config. One of `TimeBasedCanary`, `TimeBasedLinear`, `AllAtOnce`.
* `time_based_canary` - (Optional) The time based canary configuration information. If `type` is `TimeBasedLinear`, use `time_based_linear` instead.
* `time_based_linear` - (Optional) The time based linear configuration information. If `type` is `TimeBasedCanary`, use `time_based_canary` instead.

The `time_based_canary` block supports the following:

* `interval` - (Optional) The number of minutes between the first and second traffic shifts of a `TimeBasedCanary` deployment.
* `percentage` - (Optional) The percentage of traffic to shift in the first increment of a `TimeBasedCanary` deployment.

The `time_based_linear` block supports the following:

* `interval` - (Optional) The number of minutes between each incremental traffic shift of a `TimeBasedLinear` deployment.
* `percentage` - (Optional) The percentage of traffic that is shifted at the start of each increment of a `TimeBasedLinear` deployment.

The `zonal_config` block supports the following:

* `first_zone_monitor_duration_in_seconds` - (Optional) The period of time, in seconds, that CodeDeploy must wait after completing a deployment to the first Availability Zone. CodeDeploy will wait this amount of time before starting a deployment to the second Availability Zone. If you don't specify a value for `first_zone_monitor_duration_in_seconds`, then CodeDeploy uses the `monitor_duration_in_seconds` value for the first Availability Zone.
* `minimum_healthy_hosts_per_zone` - (Optional) The number or percentage of instances that must remain available per Availability Zone during a deployment. If you don't specify a value under `minimum_healthy_hosts_per_zone`, then CodeDeploy uses a default value of 0 percent. This block is more documented below.
* `monitor_duration_in_seconds` - (Optional) The period of time, in seconds, that CodeDeploy must wait after completing a deployment to an Availability Zone. CodeDeploy will wait this amount of time before starting a deployment to the next Availability Zone. If you don't specify a `monitor_duration_in_seconds`, CodeDeploy starts deploying to the next Availability Zone immediately.

The `minimum_healthy_hosts_per_zone` block supports the following:

* `type` - (Required) The type can either be `FLEET_PERCENT` or `HOST_COUNT`.
* `value` - (Required) The value when the type is `FLEET_PERCENT` represents the minimum number of healthy instances as a percentage of the total number of instances in the Availability Zone during a deployment. If you specify FLEET_PERCENT, at the start of the deployment, AWS CodeDeploy converts the percentage to the equivalent number of instance and rounds up fractional instances. When the type is `HOST_COUNT`, the value represents the minimum number of healthy instances in the Availability Zone as an absolute value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the deployment config.
* `id` - The deployment group's config name.
* `deployment_config_id` - The AWS Assigned deployment config id

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeDeploy Deployment Configurations using the `deployment_config_name`. For example:

```terraform
import {
  to = aws_codedeploy_deployment_config.example
  id = "my-deployment-config"
}
```

Using `terraform import`, import CodeDeploy Deployment Configurations using the `deployment_config_name`. For example:

```console
% terraform import aws_codedeploy_deployment_config.example my-deployment-config
```
