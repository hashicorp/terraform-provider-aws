---
layout: "aws"
page_title: "AWS: wafregional_rule"
sidebar_current: "docs-aws-resource-wafregional-rule"
description: |-
  Provides an AWS WAF Regional rule resource for use with ALB.
---

# aws\_wafregional\_rule

Provides an WAF Regional Rule Resource for use with Application Load Balancer.

## Example Usage

```hcl
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"
  
  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"
  
  predicate {
    type    = "IPMatch"
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the rule.
* `metric_name` - (Required) The name or description for the Amazon CloudWatch metric of this rule.
* `predicate` - (Optional) The ByteMatchSet, IPSet, SizeConstraintSet, SqlInjectionMatchSet, or XssMatchSet objects to include in a rule.

## Nested Fields

### `predicate`

See [docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-wafregional-rule-predicates.html)

#### Arguments

* `type` - (Required) The type of predicate in a rule, such as an IPSet (IPMatch)
* `data_id` - (Required) The unique identifier of a predicate, such as the ID of a ByteMatchSet or IPSet.
* `negated` - (Required) Whether to use the settings or the negated settings that you specified in the `ByteMatchSet`, `IPSet`, `SizeConstraintSet`, `SqlInjectionMatchSet`, or `XssMatchSet` objects.

## Remarks

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the WAF rule.
