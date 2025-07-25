---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule_group_association"
description: |-
  Associates a WAFv2 Rule Group with a Web ACL by adding a rule that references the Rule Group.
---

# Resource: aws_wafv2_web_acl_rule_group_association

Associates a WAFv2 Rule Group with a Web ACL by adding a rule that references the Rule Group. Use this resource to apply the rules defined in a Rule Group to a Web ACL without duplicating rule definitions.

~> **Note:** This resource creates a rule within the Web ACL that references the entire Rule Group. The rule group's individual rules are evaluated as a unit when requests are processed by the Web ACL.

## Example Usage

### Basic Usage

```terraform
resource "aws_wafv2_rule_group" "example" {
  name     = "example-rule-group"
  scope    = "REGIONAL"
  capacity = 10

  rule {
    name     = "block-suspicious-requests"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = ["CN", "RU"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "block-suspicious-requests"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "example-rule-group"
    sampled_requests_enabled   = true
  }
}

resource "aws_wafv2_web_acl" "example" {
  name  = "example-web-acl"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "example-web-acl"
    sampled_requests_enabled   = true
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name      = "example-rule-group-rule"
  priority       = 100
  rule_group_arn = aws_wafv2_rule_group.example.arn
  web_acl_arn    = aws_wafv2_web_acl.example.arn
}
```

### With Override Action

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name       = "example-rule-group-rule"
  priority        = 100
  rule_group_arn  = aws_wafv2_rule_group.example.arn
  web_acl_arn     = aws_wafv2_web_acl.example.arn
  override_action = "count"
}
```

### CloudFront Web ACL

```terraform
resource "aws_wafv2_rule_group" "cloudfront_example" {
  name     = "cloudfront-rule-group"
  scope    = "CLOUDFRONT"
  capacity = 10

  rule {
    name     = "rate-limit"
    priority = 1

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "rate-limit"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "cloudfront-rule-group"
    sampled_requests_enabled   = true
  }
}

resource "aws_wafv2_web_acl" "cloudfront_example" {
  name  = "cloudfront-web-acl"
  scope = "CLOUDFRONT"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "cloudfront-web-acl"
    sampled_requests_enabled   = true
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "cloudfront_example" {
  rule_name      = "cloudfront-rule-group-rule"
  priority       = 50
  rule_group_arn = aws_wafv2_rule_group.cloudfront_example.arn
  web_acl_arn    = aws_wafv2_web_acl.cloudfront_example.arn
}
```

## Argument Reference

The following arguments are required:

* `rule_name` - (Required) Name of the rule to create in the Web ACL that references the rule group. Must be between 1 and 128 characters.
* `priority` - (Required) Priority of the rule within the Web ACL. Rules are evaluated in order of priority, with lower numbers evaluated first.
* `rule_group_arn` - (Required) ARN of the Rule Group to associate with the Web ACL.
* `web_acl_arn` - (Required) ARN of the Web ACL to associate the Rule Group with.

The following arguments are optional:

* `override_action` - (Optional) Override action for the rule group. Valid values are `none` and `count`. Defaults to `none`. When set to `count`, the actions defined in the rule group rules are overridden to count matches instead of blocking or allowing requests.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Rule Group Associations using `WebACLARN,RuleGroupARN,RuleName`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule_group_association.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321,example-rule-group-rule"
}
```

Using `terraform import`, import WAFv2 Web ACL Rule Group Associations using `WebACLARN,RuleGroupARN,RuleName`. For example:

```console
% terraform import aws_wafv2_web_acl_rule_group_association.example "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321,example-rule-group-rule"
```
