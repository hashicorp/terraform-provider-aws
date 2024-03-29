# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "rule_arn" {
  value = aws_cloudwatch_event_rule.foo.arn
}

output "sns_topic_arn" {
  value = aws_sns_topic.foo.arn
}
