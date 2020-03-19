---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_subscribed_rule_group"
description: |-
  Retrieves an AWS WAF Regional Subscribed Rule Group id.
---

# Data Source: aws_wafregional_subscribed_rule_group

`aws_wafregional_subscribed_rule_group` Retrieves a WAF Regional Subscribed Rule Group Resource Id.

## Example Usage

```hcl
data "aws_wafregional_subscribed_rule_group" "example" {
  name = "tfWAFRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF Regional subscribed rule group.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional subscribed rule group.
* `metric_name` - The Metric Name of the WAF Regional subscribed rule group.
