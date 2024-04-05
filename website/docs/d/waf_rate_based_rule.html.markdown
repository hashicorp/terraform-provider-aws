---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_rate_based_rule"
description: |-
  Retrieves an AWS WAF rate based rule id.
---

# Data Source: aws_waf_rate_based_rule

`aws_waf_rate_based_rule` Retrieves a WAF Rate Based Rule Resource Id.

## Example Usage

```terraform
data "aws_waf_rate_based_rule" "example" {
  name = "tfWAFRateBasedRule"
}

```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAF rate based rule.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF rate based rule.
