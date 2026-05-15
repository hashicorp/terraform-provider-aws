# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = var.rName

  rule {
    resource_type  = "AWS::CloudWatch::OTelEnrichment"
    telemetry_type = "Logs"
  }
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
