---
subcategory: "ARC (Application Recovery Controller) Zonal Shift"
layout: "aws"
page_title: "AWS: aws_arczonalshift_zonal_autoshift_configuration"
description: |-
  Manages an AWS Application Recovery Controller Zonal Shift Zonal Autoshift Configuration.
---

# Resource: aws_arczonalshift_zonal_autoshift_configuration

Manages an AWS Application Recovery Controller Zonal Shift Zonal Autoshift Configuration for a managed resource (such as a load balancer).

Zonal autoshift is a capability in AWS Application Recovery Controller (ARC) that automatically shifts traffic away from an Availability Zone when AWS identifies a potential issue, helping maintain application availability.

## Example Usage

### Basic Usage

```terraform
resource "aws_arczonalshift_zonal_autoshift_configuration" "example" {
  resource_arn           = aws_lb.example.arn
  zonal_autoshift_status = "ENABLED"

  outcome_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.example.arn
    type             = "CLOUDWATCH"
  }
}

resource "aws_lb" "example" {
  name               = "example"
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.example[*].id

  enable_zonal_shift = true
}

resource "aws_cloudwatch_metric_alarm" "example" {
  alarm_name          = "example-outcome-alarm"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Outcome alarm for zonal autoshift practice run"

  dimensions = {
    LoadBalancer = aws_lb.example.arn_suffix
  }
}
```

### With Blocking Alarms

```terraform
resource "aws_arczonalshift_zonal_autoshift_configuration" "example" {
  resource_arn           = aws_lb.example.arn
  zonal_autoshift_status = "ENABLED"

  outcome_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.outcome.arn
    type             = "CLOUDWATCH"
  }

  blocking_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.blocking.arn
    type             = "CLOUDWATCH"
  }
}
```

### With Blocked Windows

```terraform
resource "aws_arczonalshift_zonal_autoshift_configuration" "example" {
  resource_arn           = aws_lb.example.arn
  zonal_autoshift_status = "ENABLED"
  blocked_windows        = ["Mon:00:00-Mon:08:00"]

  outcome_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.example.arn
    type             = "CLOUDWATCH"
  }
}
```

## Argument Reference

The following arguments are required:

* `outcome_alarms` - (Required) List of CloudWatch alarms monitored during practice runs. See [`outcome_alarms`](#outcome_alarms) below.
* `resource_arn` - (Required) The ARN of the managed resource to configure zonal autoshift for (e.g., an Application Load Balancer). Changing this creates a new resource.
* `zonal_autoshift_status` - (Required) The status of zonal autoshift. Valid values: `ENABLED`, `DISABLED`.

The following arguments are optional:

* `allowed_windows` - (Optional) List of time windows during which practice runs are allowed, in the format `Day:HH:MM-Day:HH:MM` (e.g., `Mon:09:00-Mon:17:00`). Cannot be used together with `blocked_windows`.
* `blocked_dates` - (Optional) List of dates when practice runs should not be started, in the format `YYYY-MM-DD`.
* `blocked_windows` - (Optional) List of time windows during which practice runs should not be started, in the format `Day:HH:MM-Day:HH:MM` (e.g., `Mon:00:00-Mon:08:00`). Cannot be used together with `allowed_windows`.
* `blocking_alarms` - (Optional) List of CloudWatch alarms that can block practice runs when in alarm state. See [`blocking_alarms`](#blocking_alarms) below.
* `region` - (Optional) AWS region where the resource is deployed.

### `outcome_alarms`

* `alarm_identifier` - (Required) ARN of the CloudWatch alarm.
* `type` - (Required) Type of control condition. Valid value: `CLOUDWATCH`.

### `blocking_alarms`

* `alarm_identifier` - (Required) ARN of the CloudWatch alarm.
* `type` - (Required) Type of control condition. Valid value: `CLOUDWATCH`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_arczonalshift_zonal_autoshift_configuration.example
  identity = {
    resource_arn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/example/50dc6c495c0c9188"
  }
}

resource "aws_arczonalshift_zonal_autoshift_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the managed resource to configure zonal autoshift for.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ARC Zonal Shift Zonal Autoshift Configuration using the `resource_arn`. For example:

```terraform
import {
  to = aws_arczonalshift_zonal_autoshift_configuration.example
  id = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/example/50dc6c495c0c9188"
}
```

Using `terraform import`, import ARC Zonal Shift Zonal Autoshift Configuration using the `resource_arn`. For example:

```console
% terraform import aws_arczonalshift_zonal_autoshift_configuration.example arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/example/50dc6c495c0c9188
```
