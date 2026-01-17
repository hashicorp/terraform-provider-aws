resource "aws_cloudwatch_contributor_insight_rule" "test" {
  rule_name       = var.rName
  rule_state      = "ENABLED"
  rule_definition = "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"

  {{- template "tags" . }}
}