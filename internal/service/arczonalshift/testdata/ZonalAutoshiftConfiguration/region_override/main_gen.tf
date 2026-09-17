# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  region = var.region

  resource_arn           = aws_lb.test.arn
  zonal_autoshift_status = "ENABLED"

  outcome_alarms {
    alarm_identifier = aws_cloudwatch_metric_alarm.outcome.arn
    type             = "CLOUDWATCH"
  }
}

resource "aws_lb" "test" {
  region = var.region

  name               = var.rName
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
  enable_zonal_shift         = true

  tags = {
    Name = var.rName
  }
}

resource "aws_cloudwatch_metric_alarm" "outcome" {
  region = var.region

  alarm_name          = "${var.rName}-outcome"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Outcome alarm for zonal autoshift practice run"

  dimensions = {
    LoadBalancer = aws_lb.test.arn_suffix
  }
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 2)

resource "aws_subnet" "test" {
  region = var.region

  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  region = var.region

  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
