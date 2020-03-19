---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_waf_subscribed_rule_group"
description: |-
  Retrieves an AWS WAF Subscribed Rule Group id.
---

# Data Source: aws_waf_subscribed_rule_group

`aws_waf_subscribed_rule_group` Retrieves a WAF Subscribed Rule Group Resource Id.

## Example Usage

```hcl
data "aws_waf_subscribed_rule_group" "example" {
  name = "tfWAFRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF subscribed rule group.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF subscribed rule group.
* `metric_name` - The Metric Name of the WAF subscribed rule group.
