---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_record"
description: |-
  Lists Route 53 Record resources.
---

# List Resource: aws_route53_record

Lists Route 53 Record resources.

## Example Usage

```terraform
list "aws_route53_record" "example" {
  provider = aws
  config {
    zone_id = aws_route53_zone.example.zone_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `zone_id` - (Required) ID of the hosted zone to list records from.
