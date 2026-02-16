# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "rule_arn" {
  value = aws_cloudwatch_event_rule.foo.arn
}

output "sns_topic_arn" {
  value = aws_sns_topic.foo.arn
}
