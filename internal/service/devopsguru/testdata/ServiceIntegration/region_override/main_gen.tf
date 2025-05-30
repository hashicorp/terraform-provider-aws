# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsguru_service_integration" "test" {
  region = var.region

  # Default to existing configured settings
  kms_server_side_encryption {}

  logs_anomaly_detection {
    opt_in_status = "DISABLED"
  }
  ops_center {
    opt_in_status = "DISABLED"
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
