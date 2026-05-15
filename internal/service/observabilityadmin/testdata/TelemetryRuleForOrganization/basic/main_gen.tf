# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = var.rName

  rule {
    resource_type  = "AWS::SecurityHub::HubV2"
    telemetry_type = "Logs"
  }
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
