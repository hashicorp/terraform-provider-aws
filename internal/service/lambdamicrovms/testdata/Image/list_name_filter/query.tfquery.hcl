# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lambdamicrovms_image" "test" {
  provider = aws

  config {
    name_filter = "${var.rName}-0"
  }
}
