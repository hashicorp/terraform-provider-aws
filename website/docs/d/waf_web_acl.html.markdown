---
layout: "aws"
page_title: "AWS: aws_waf_web_acl"
sidebar_current: "docs-aws-datasource-waf-web-acl"
description: |-
  Retrieves a WAF Web ACL id.
---

# Data Source: aws_waf_rule

`aws_waf_rule` Retrieves a WAF Web ACL Resource Id.

## Example Usage

```hcl
data "aws_waf_web_acl" "example" {
  name = "tfWAFRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF Web ACL.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF WebACL.