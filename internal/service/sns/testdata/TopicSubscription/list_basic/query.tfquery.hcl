# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_sns_topic_subscription" "test" {
  provider = aws

  config {
    topic_arn = aws_sns_topic.test.arn
  }
}
