# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_rule" "test" {
  region = var.region


  rule_name = var.rName

  rule {
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.test]
}

resource "aws_observabilityadmin_telemetry_evaluation" "test" {
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
