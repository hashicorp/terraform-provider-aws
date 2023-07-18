---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl"
description: |-
  Retrieves a WAF Regional Web ACL id.
---

# Data Source: aws_wafregional_web_acl

`aws_wafregional_web_acl` Retrieves a WAF Regional Web ACL Resource Id.

## Example Usage

```terraform
data "aws_wafregional_web_acl" "example" {
  name = "tfWAFRegionalWebACL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the WAF Regional Web ACL.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Regional Web ACL.
