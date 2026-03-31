---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_route_table_association"
description: |-
  Lists EC2 (Elastic Compute Cloud) Route Table Association resources.
---

# List Resource: aws_ec2_route_table_association

Lists EC2 (Elastic Compute Cloud) Route Table Association resources.

## Example Usage

```terraform
list "aws_ec2_route_table_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
