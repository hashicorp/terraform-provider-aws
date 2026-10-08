# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_prometheus_anomaly_detector" "first" {
  alias        = "${var.rName}-first"
  workspace_id = aws_prometheus_workspace.first.id

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  missing_data_action {
    skip = true
  }
}

resource "aws_prometheus_anomaly_detector" "second" {
  alias        = "${var.rName}-second"
  workspace_id = aws_prometheus_workspace.second.id

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  missing_data_action {
    skip = true
  }
}

resource "aws_prometheus_workspace" "first" {}

resource "aws_prometheus_workspace" "second" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}