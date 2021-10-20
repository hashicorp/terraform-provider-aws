output "rule_arn" {
  value = aws_cloudwatchevents_rule.foo.arn
}

output "kinesis_stream_arn" {
  value = aws_kinesis_stream.foo.arn
}
