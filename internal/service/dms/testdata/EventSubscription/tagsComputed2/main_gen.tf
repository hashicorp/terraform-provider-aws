# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_dms_event_subscription" "test" {
  name             = var.rName
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

# testAccEventSubscriptionConfig_base

data "aws_partition" "current" {}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = var.rName
  replication_subnet_group_id          = var.rName
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id     = var.rName
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
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

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
