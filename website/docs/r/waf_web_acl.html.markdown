---
layout: "aws"
page_title: "AWS: aws_waf_web_acl"
sidebar_current: "docs-aws-resource-waf-webacl"
description: |-
  Provides a AWS WAF web access control group (ACL) resource.
---

# aws_waf_web_acl

Provides a WAF Web ACL Resource

## Example Usage

```hcl
resource "aws_waf_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = ["aws_waf_ipset.ipset"]
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicates {
    data_id = "${aws_waf_ipset.ipset.id}"
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "waf_acl" {
  depends_on  = ["aws_waf_ipset.ipset", "aws_waf_rule.wafrule"]
  name        = "tfWebACL"
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }

  rules {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = "${aws_waf_rule.wafrule.id}"
    type     = "REGULAR"
  }
}
```

## Argument Reference

The following arguments are supported:

* `default_action` - (Required) The action that you want AWS WAF to take when a request doesn't match the criteria in any of the rules that are associated with the web ACL.
* `metric_name` - (Required) The name or description for the Amazon CloudWatch metric of this web ACL.
* `name` - (Required) The name or description of the web ACL.
* `rules` - (Required) The rules to associate with the web ACL and the settings for each rule.

## Nested Blocks

### `default_action`

#### Arguments

* `type` - (Required) Specifies how you want AWS WAF to respond to requests that match the settings in a rule.
  e.g. `ALLOW`, `BLOCK` or `COUNT`

### `rules`

See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_ActivatedRule.html) for all details and supported values.

#### Arguments

* `action` - (Optional) The action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Not used if `type` is `GROUP`.
  * `type` - (Required) valid values are: `BLOCK`, `ALLOW`, or `COUNT`
* `override_action` - (Optional) Override the action that a group requests CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Only used if `type` is `GROUP`.
  * `type` - (Required) valid values are: `NONE` or `COUNT`
* `priority` - (Required) Specifies the order in which the rules in a WebACL are evaluated.
  Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) ID of the associated WAF (Global) rule (e.g. [`aws_waf_rule`](/docs/providers/aws/r/waf_rule.html)). WAF (Regional) rules cannot be used.
* `type` - (Optional) The rule type, either `REGULAR`, as defined by [Rule](http://docs.aws.amazon.com/waf/latest/APIReference/API_Rule.html), `RATE_BASED`, as defined by [RateBasedRule](http://docs.aws.amazon.com/waf/latest/APIReference/API_RateBasedRule.html), or `GROUP`, as defined by [RuleGroup](https://docs.aws.amazon.com/waf/latest/APIReference/API_RuleGroup.html). The default is REGULAR. If you add a RATE_BASED rule, you need to set `type` as `RATE_BASED`. If you add a GROUP rule, you need to set `type` as `GROUP`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF WebACL.

## Import

WAF Web ACL can be imported using the `id`, e.g.

```
$ terraform import aws_waf_web_acl.main 0c8e583e-18f3-4c13-9e2a-67c4805d2f94
```
