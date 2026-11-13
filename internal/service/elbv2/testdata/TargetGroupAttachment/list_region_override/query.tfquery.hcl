# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lb_target_group_attachment" "test" {
  provider = aws

  config {
    region           = var.region
    target_group_arn = aws_lb_target_group.test.arn
  }
}
