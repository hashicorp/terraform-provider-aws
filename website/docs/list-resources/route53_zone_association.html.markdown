---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zone_association"
description: |-
  Lists Route 53 Zone Association resources.
---

# List Resource: aws_route53_zone_association

Lists Route 53 Zone Association resources.

## Example Usage

```terraform
list "aws_route53_zone_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
