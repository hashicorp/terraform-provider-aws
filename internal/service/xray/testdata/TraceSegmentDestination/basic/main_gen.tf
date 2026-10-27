# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_trace_segment_destination" "test" {
  destination = "XRay"
}

