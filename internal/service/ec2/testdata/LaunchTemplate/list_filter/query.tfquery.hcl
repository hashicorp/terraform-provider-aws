# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_launch_template" "test" {
  provider = aws

  config {
    launch_template_ids = [aws_launch_template.expected[0].id, aws_launch_template.expected[1].id]
  }
}
