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

!> **Warning:** Using this resource will cause the associated Web ACL resource to show configuration drift in the `rule` argument unless you add `lifecycle { ignore_changes = [rule] }` to the Web ACL resource configuration. This is because this resource modifies the Web ACL's rules outside of the Web ACL resource's direct management.

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

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "example-rule-group-rule"
  priority    = 100
  web_acl_arn = aws_wafv2_web_acl.example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
  }
}
```

### With Override Action

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name       = "example-rule-group-rule"
  priority        = 100
  web_acl_arn     = aws_wafv2_web_acl.example.arn
  override_action = "count"

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
  }
}
```

### With Rule Action Overrides

```terraform
resource "aws_wafv2_rule_group" "example" {
  name     = "example-rule-group"
  scope    = "REGIONAL"
  capacity = 10

  rule {
    name     = "geo-block-rule"
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
      metric_name                = "geo-block-rule"
      sampled_requests_enabled   = true
    }
  }

  rule {
    name     = "rate-limit-rule"
    priority = 2

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 1000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "rate-limit-rule"
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

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "example-rule-group-rule"
  priority    = 100
  web_acl_arn = aws_wafv2_web_acl.example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
    
    # Override specific rules within the rule group
    rule_action_override {
      name = "geo-block-rule"
      action_to_use {
        count {
          custom_request_handling {
            insert_header {
              name  = "X-Geo-Block-Override"
              value = "counted"
            }
          }
        }
      }
    }

    rule_action_override {
      name = "rate-limit-rule"
      action_to_use {
        captcha {
          custom_request_handling {
            insert_header {
              name  = "X-Rate-Limit-Override"
              value = "captcha-required"
            }
          }
        }
      }
    }
  }
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

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "cloudfront_example" {
  rule_name   = "cloudfront-rule-group-rule"
  priority    = 50
  web_acl_arn = aws_wafv2_web_acl.cloudfront_example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.cloudfront_example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `rule_name` - (Required) Name of the rule to create in the Web ACL that references the rule group. Must be between 1 and 128 characters.
* `priority` - (Required) Priority of the rule within the Web ACL. Rules are evaluated in order of priority, with lower numbers evaluated first.
* `web_acl_arn` - (Required) ARN of the Web ACL to associate the Rule Group with.
* `rule_group_reference` - (Required) Rule Group reference configuration. [See below](#rule_group_reference).

The following arguments are optional:

* `override_action` - (Optional) Override action for the rule group. Valid values are `none` and `count`. Defaults to `none`. When set to `count`, the actions defined in the rule group rules are overridden to count matches instead of blocking or allowing requests.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### rule_group_reference

* `arn` - (Required) ARN of the Rule Group to associate with the Web ACL.
* `rule_action_override` - (Optional) Override actions for specific rules within the rule group. [See below](#rule_action_override).

### rule_action_override

* `name` - (Required) Name of the rule to override within the rule group.
* `action_to_use` - (Required) Action to use instead of the rule's original action. [See below](#action_to_use).

### action_to_use

Exactly one of the following action blocks must be specified:

* `allow` - (Optional) Allow the request. [See below](#allow).
* `block` - (Optional) Block the request. [See below](#block).
* `captcha` - (Optional) Require CAPTCHA verification. [See below](#captcha).
* `challenge` - (Optional) Require challenge verification. [See below](#challenge).
* `count` - (Optional) Count the request without taking action. [See below](#count).

### allow

* `custom_request_handling` - (Optional) Custom handling for allowed requests. [See below](#custom_request_handling).

### block

* `custom_response` - (Optional) Custom response for blocked requests. [See below](#custom_response).

### captcha

* `custom_request_handling` - (Optional) Custom handling for CAPTCHA requests. [See below](#custom_request_handling).

### challenge

* `custom_request_handling` - (Optional) Custom handling for challenge requests. [See below](#custom_request_handling).

### count

* `custom_request_handling` - (Optional) Custom handling for counted requests. [See below](#custom_request_handling).

### custom_request_handling

* `insert_header` - (Required) Headers to insert into the request. [See below](#insert_header).

### custom_response

* `custom_response_body_key` - (Optional) Key of a custom response body to use.
* `response_code` - (Required) HTTP response code to return (200-599).
* `response_header` - (Optional) Headers to include in the response. [See below](#response_header).

### insert_header

* `name` - (Required) Name of the header to insert.
* `value` - (Required) Value of the header to insert.

### response_header

* `name` - (Required) Name of the response header.
* `value` - (Required) Value of the response header.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Rule Group Associations using `WebACLARN,RuleName,RuleGroupARN`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule_group_association.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,example-rule-group-rule,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321"
}
```

Using `terraform import`, import WAFv2 Web ACL Rule Group Associations using `WebACLARN,RuleName,RuleGroupARN`. For example:

```console
% terraform import aws_wafv2_web_acl_rule_group_association.example "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,example-rule-group-rule,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321"
```
