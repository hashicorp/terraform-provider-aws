# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_prometheus_anomaly_detector" "test" {
  count = var.resource_count

  alias        = "${var.rName}-${count.index}"
  workspace_id = aws_prometheus_workspace.test.id

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  missing_data_action {
    skip = true
  }
}

resource "aws_prometheus_workspace" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
