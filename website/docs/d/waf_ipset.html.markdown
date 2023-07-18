---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_ipset"
description: |-
  Retrieves an AWS WAF IP set id.
---

# Data Source: aws_waf_ipset

`aws_waf_ipset` Retrieves a WAF IP Set Resource Id.

## Example Usage

```terraform
data "aws_waf_ipset" "example" {
  name = "tfWAFIPSet"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the WAF IP set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF IP set.
