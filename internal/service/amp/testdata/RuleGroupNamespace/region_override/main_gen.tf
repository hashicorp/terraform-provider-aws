# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_prometheus_rule_group_namespace" "test" {
  region = var.region

  name         = var.rName
  workspace_id = aws_prometheus_workspace.test.id
  data         = <<EOF
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
EOF
}

resource "aws_prometheus_workspace" "test" {
  region = var.region

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
