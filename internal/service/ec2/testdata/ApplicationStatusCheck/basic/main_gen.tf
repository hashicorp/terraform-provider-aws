# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80
}
