---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_rule_group_permission_policy"
description: |-
  Attaches a permission policy to a WAFv2 Rule Group to share it with other AWS accounts.
---

# Resource: aws_wafv2_rule_group_permission_policy

Attaches a permission policy to a WAFv2 Rule Group, enabling cross-account sharing.
The policy allows specified AWS accounts to reference the rule group in their web ACLs.

For more information, see [Sharing a rule group](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-group-sharing.html) in the AWS WAF Developer Guide.

## Example Usage

### Share with a specific account

```terraform
resource "aws_wafv2_rule_group" "example" {
  name     = "example"
  scope    = "REGIONAL"
  capacity = 10

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "example"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_rule_group_permission_policy" "example" {
  resource_arn = aws_wafv2_rule_group.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:aws:iam::111111111111:root"
      }
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
        "wafv2:GetRuleGroup",
      ]
    }]
  })
}
```

### Share with multiple accounts

```terraform
resource "aws_wafv2_rule_group_permission_policy" "example" {
  resource_arn = aws_wafv2_rule_group.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = [
          "arn:aws:iam::111111111111:root",
          "arn:aws:iam::222222222222:root",
        ]
      }
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
        "wafv2:GetRuleGroup",
      ]
    }]
  })
}
```

### Share with all accounts in an organization

```terraform
resource "aws_wafv2_rule_group_permission_policy" "example" {
  resource_arn = aws_wafv2_rule_group.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
        "wafv2:GetRuleGroup",
      ]
      Condition = {
        StringEquals = {
          "aws:PrincipalOrgID" = "o-example1234"
        }
      }
    }]
  })
}
```

### Share with a specific organizational unit (OU)

```terraform
resource "aws_wafv2_rule_group_permission_policy" "example" {
  resource_arn = aws_wafv2_rule_group.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "wafv2:CreateWebACL",
        "wafv2:UpdateWebACL",
        "wafv2:PutFirewallManagerRuleGroups",
        "wafv2:GetRuleGroup",
      ]
      Condition = {
        "ForAnyValue:StringLike" = {
          "aws:PrincipalOrgPaths" = "o-example1234/r-ab12/ou-ab12-example/*"
        }
      }
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) The IAM policy to attach to the rule group. The policy must conform to the following:
    * IAM Policy version must be `2012-10-17`.
    * Must include specifications for `Effect`, `Action`, and `Principal`.
    * `Effect` must be `Allow`.
    * `Action` must include `wafv2:CreateWebACL`, `wafv2:UpdateWebACL`, and `wafv2:PutFirewallManagerRuleGroups`. May optionally include `wafv2:GetRuleGroup`. AWS WAF rejects any extra actions or wildcard actions.
    * Must not include a `Resource` parameter.
* `resource_arn` - (Required, Forces new resource) The ARN of the WAFv2 Rule Group to attach the policy to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the WAFv2 Rule Group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a WAFv2 Rule Group Permission Policy using the rule group ARN. For example:

```terraform
import {
  to = aws_wafv2_rule_group_permission_policy.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"
}
```

Using `terraform import`, import a WAFv2 Rule Group Permission Policy using the rule group ARN. For example:

```console
% terraform import aws_wafv2_rule_group_permission_policy.example arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111
```
