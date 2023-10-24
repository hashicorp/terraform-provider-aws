---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_ipset"
description: |-
  Retrieves an AWS WAF Regional IP set id.
---

# Data Source: aws_wafregional_ipset

`aws_wafregional_ipset` Retrieves a WAF Regional IP Set Resource Id.

## Example Usage

```terraform
data "aws_wafregional_ipset" "example" {
  name = "tfWAFRegionalIPSet"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAF Regional IP set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Regional IP set.
