---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_web_acl"
description: |-
  Retrieves a WAF Web ACL id.
---

# Data Source: aws_waf_web_acl

`aws_waf_web_acl` Retrieves a WAF Web ACL Resource ID.

## Example Usage

```terraform
data "aws_waf_web_acl" "example" {
  name = "tfWAFWebACL"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Name of the WAF Web ACL.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Web ACL.
