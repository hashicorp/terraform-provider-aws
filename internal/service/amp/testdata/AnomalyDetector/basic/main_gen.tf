# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_prometheus_workspace" "test" {
}

resource "aws_prometheus_anomaly_detector" "test" {
  alias = var.rName
  workspace_id = aws_prometheus_workspace.test.id

  configuration {
	random_cut_forest {
	  query = "avg(up)"
	}
  }

  missing_data_action{
    skip = true
  }
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
