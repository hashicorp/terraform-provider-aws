# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_contributor_managed_insight_rule" "test" {
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = "VpcEndpointService-NewConnectionsByEndpointId-v1"

  tags = var.resource_tags
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 2)

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

resource "aws_lb" "test" {
  count = 2

  load_balancer_type = "network"
  name               = "${var.rName}-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
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
