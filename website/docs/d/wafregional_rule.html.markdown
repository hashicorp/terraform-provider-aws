---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_rule"
description: |-
  Retrieves an AWS WAF Regional rule id.
---

# Data Source: aws_wafregional_rule

`aws_wafregional_rule` Retrieves a WAF Regional Rule Resource Id.

## Example Usage

```terraform
data "aws_wafregional_rule" "example" {
  name = "tfWAFRegionalRule"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAF Regional rule.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Regional rule.
