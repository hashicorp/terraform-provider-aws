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

  config {
    vpc_id = aws_vpc.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `vpc_id` - (Required) ID of the VPC to list hosted zone associations for.
* `vpc_region` - (Optional) Region of the VPC. Defaults to the provider region.
