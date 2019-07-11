---
layout: "aws"
page_title: "AWS: aws_wafregional_rule"
sidebar_current: "docs-aws-datasource-wafregional-rule"
description: |-
  Retrieves an AWS WAF Regional rule id.
---

# Data Source: aws_waf_rule

`aws_wafregional_rule` Retrieves a WAF Regional Rule Resource Id.

## Example Usage

```hcl
data "aws_wafregional_rule" "example" {
  name = "tfWAFRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF rule.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional rule.
