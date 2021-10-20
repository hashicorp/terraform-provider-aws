output "rule_arn" {
  value = aws_cloudwatchevents_rule.foo.arn
}

output "sns_topic_arn" {
  value = aws_sns_topic.foo.arn
}
