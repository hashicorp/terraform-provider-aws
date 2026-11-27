---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_vpc_association_authorization"
description: |-
  Lists Route 53 VPC Association Authorization resources.
---

# List Resource: aws_route53_vpc_association_authorization

Lists Route 53 VPC Association Authorization resources.

## Example Usage

```terraform
list "aws_route53_vpc_association_authorization" "example" {
  provider = aws

  config {
    zone_id = "Z1234567890"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `zone_id` - (Required) ID of the hosted zone to list VPC association authorizations for.
