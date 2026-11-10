---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_ip_set"
description: |-
  Lists WAFv2 IP Set resources.
---

# List Resource: aws_wafv2_ip_set

Lists WAFv2 IP Set resources.

## Example Usage

```terraform
list "aws_wafv2_ip_set" "example" {
  provider = aws

  config {
    scope = "REGIONAL"
  }
}
```

## Argument Reference

The following arguments are required:

* `scope` - (Required) Whether to list IP Sets for a global (`CLOUDFRONT`) or regional (`REGIONAL`) application. Valid values are `CLOUDFRONT` and `REGIONAL`. To list `CLOUDFRONT` scoped IP Sets, set the `region` to `us-east-1`.

The following arguments are optional:

* `region` - (Optional) Region to query. Defaults to provider region.
