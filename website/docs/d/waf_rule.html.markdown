---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_rule"
description: |-
  Retrieves an AWS WAF rule id.
---

# Data Source: aws_waf_rule

`aws_waf_rule` Retrieves a WAF Rule Resource ID.

## Example Usage

```terraform
data "aws_waf_rule" "example" {
  name = "tfWAFRule"
}

```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Name of the WAF rule.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF rule.
