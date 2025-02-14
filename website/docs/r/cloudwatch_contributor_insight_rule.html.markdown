---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_contributor_insight_rule"
description: |-
  Terraform resource for managing an AWS CloudWatch Contributor Insight Rule.
---

# Resource: aws_cloudwatch_contributor_insight_rule

Terraform resource for managing an AWS CloudWatch Contributor Insight Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_contributor_insight_rule" "test" {
  rule_name       = "testing"
  rule_state      = "ENABLED"
  rule_definition = "{\"Schema\":{\"Name\":\"CloudWatchLogRule\",\"Version\":1},\"AggregateOn\":\"Count\",\"Contribution\":{\"Filters\":[{\"In\":[\"some-keyword\"],\"Match\":\"$.message\"}],\"Keys\":[\"$.country\"]},\"LogFormat\":\"JSON\",\"LogGroupNames\":[\"/aws/lambda/api-prod\"]}"
}
```

## Argument Reference

The following arguments are required:

* `rule_definition` - (Required) Definition of the rule, as a JSON object. For details on the valid syntax, see [Contributor Insights Rule Syntax](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/ContributorInsights-RuleSyntax.html).
* `rule_name` - (Required) Unique name of the rule.

The following arguments are optional:

* `rule_state` - (Optional) State of the rule. Valid values are `ENABLED` and `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `resource_arn` - ARN of the Contributor Insight Rule.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Contributor Insight Rule using the `rule_name`. For example:

```terraform
import {
  to = aws_cloudwatch_contributor_insight_rule.example
  id = "contributor_insight_rule-name"
}
```

Using `terraform import`, import CloudWatch Contributor Insight Rule using the `rule_name`. For example:

```console
% terraform import aws_cloudwatch_contributor_insight_rule.example contributor_insight_rule-name
```
