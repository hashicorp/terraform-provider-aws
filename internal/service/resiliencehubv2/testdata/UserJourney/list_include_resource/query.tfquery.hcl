# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_resiliencehubv2_user_journey" "test" {
  provider = aws

  include_resource = true

  config {
    system_arn = aws_resiliencehubv2_system.test.arn
  }
}
