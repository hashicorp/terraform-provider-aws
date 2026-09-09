# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_resiliencehubv2_assertion" "test" {
  provider = aws

  config {
    service_arn = aws_resiliencehubv2_service.test.arn
  }
}
