# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = var.rName
  alarm_rule = join(" OR ", formatlist("ALARM(%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))

  tags = var.resource_tags
}

resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "${var.rName}-${count.index}"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
