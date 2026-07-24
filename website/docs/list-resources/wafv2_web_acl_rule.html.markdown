---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule"
description: |-
  Lists WAFv2 Web ACL Rule resources.
---

# List Resource: aws_wafv2_web_acl_rule

Lists WAFv2 Web ACL Rule resources.

## Example Usage

```terraform
list "aws_wafv2_web_acl_rule" "example" {
  provider = aws

  config {
    web_acl_arn = aws_wafv2_web_acl.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `web_acl_arn` - (Required) ARN of the Web ACL whose rules to list.
