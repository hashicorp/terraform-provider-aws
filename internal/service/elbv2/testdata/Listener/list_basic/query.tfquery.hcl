# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lb_listener" "test" {
  provider = aws

  config {
    load_balancer_arn = aws_lb.test.arn
  }
}
