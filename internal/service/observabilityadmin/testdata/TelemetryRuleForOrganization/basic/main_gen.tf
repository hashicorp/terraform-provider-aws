# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = var.rName

  rule {
    telemetry_type = "Metrics"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.test]
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
