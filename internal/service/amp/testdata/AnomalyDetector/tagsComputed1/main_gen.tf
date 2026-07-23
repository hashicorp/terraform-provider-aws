# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
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
