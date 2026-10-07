# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_prometheus_anomaly_detector" "test" {
  provider = aws

  include_resource = true

  config {
    workspace_id = aws_prometheus_workspace.test.id
  }
}
