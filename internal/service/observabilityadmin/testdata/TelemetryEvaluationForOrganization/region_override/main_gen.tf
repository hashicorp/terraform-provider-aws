# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {
  region = var.region

}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
