---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_route_table_association"
description: |-
  Lists VPC Endpoint Route Table Association resources.
---

# List Resource: aws_vpc_endpoint_route_table_association

Lists VPC Endpoint Route Table Association resources.

## Example Usage

```terraform
list "aws_vpc_endpoint_route_table_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
