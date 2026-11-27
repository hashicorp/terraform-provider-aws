---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule_association"
description: |-
  Lists Route 53 Resolver Rule Association resources.
---

# List Resource: aws_route53_resolver_rule_association

Lists Route 53 Resolver Rule Association resources.

## Example Usage

```terraform
list "aws_route53_resolver_rule_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
