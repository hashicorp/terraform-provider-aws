# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lb_listener_rule" "test" {
  provider = aws

  config {
    region       = var.region
    listener_arn = aws_lb_listener.test.arn
  }
}
