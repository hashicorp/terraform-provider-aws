# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_sampling_rule" "test" {
  rule_name      = var.rName
  priority       = var.priority
  reservoir_size = var.reservoir_size
  url_path       = "*"
  host           = "*"
  http_method    = "GET"
  service_type   = "*"
  service_name   = "*"
  fixed_rate     = 0.3
  resource_arn   = "*"
  version        = 1
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "priority" {
  description = "Priority"
  type        = number
  nullable    = false
}

variable "reservoir_size" {
  description = "Reservoir size"
  type        = number
  nullable    = false
}
