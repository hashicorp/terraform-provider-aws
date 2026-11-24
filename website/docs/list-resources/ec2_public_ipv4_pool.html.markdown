---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_public_ipv4_pool"
description: |-
  Lists EC2 Public IPv4 Pool resources.
---

# List Resource: aws_ec2_public_ipv4_pool

Lists EC2 Public IPv4 Pool resources.

## Example Usage

```terraform
list "aws_ec2_public_ipv4_pool" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
