# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  count = var.resource_count

  rule_name = "${var.rName}-${count.index}"
  rule {
    resource_type  = count.index == 0 ? "AWS::SecurityHub::Hub" : "AWS::MSK::Cluster"
    telemetry_type = "Logs"
  }

  tags = var.resource_tags

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.test]
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {}

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

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
