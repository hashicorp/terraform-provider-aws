# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_contributor_insight_rule" "test" {
  rule_name       = var.rName
  rule_state      = "ENABLED"
  rule_definition = "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
