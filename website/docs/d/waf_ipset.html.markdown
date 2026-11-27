---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_ipset"
description: |-
  Retrieves an AWS WAF IP set id.
---

# Data Source: aws_waf_ipset

`aws_waf_ipset` Retrieves a WAF IP Set Resource ID.

## Example Usage

```terraform
data "aws_waf_ipset" "example" {
  name = "tfWAFIPSet"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Name of the WAF IP set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF IP set.
