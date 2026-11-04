# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_arns     = ["arn:${data.aws_partition.current.partition}:network-security-manager:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:rule:tfacc000000000000000000001"]
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
