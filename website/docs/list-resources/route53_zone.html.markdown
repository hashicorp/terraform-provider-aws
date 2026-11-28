---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zone"
description: |-
  Lists Route 53 Zone resources.
---

# List Resource: aws_route53_zone

Lists Route 53 Zone resources.

## Example Usage

```terraform
list "aws_route53_zone" "example" {
  provider = aws
}
```

```terraform
list "aws_route53_zone" "private" {
  provider = aws

  config {
    private_zone = true
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `private_zone` - (Optional) When `true`, only private hosted zones are returned. If omitted or `false`, both public and private hosted zones are returned.
