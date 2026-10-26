# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_trace_segment_destination" "test" {
  region = var.region

  destination = "XRay"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
