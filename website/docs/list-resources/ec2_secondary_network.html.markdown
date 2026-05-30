---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_secondary_network"
description: |-
  Lists EC2 (Elastic Compute Cloud) Secondary Network resources.
---

# List Resource: aws_ec2_secondary_network

Lists EC2 (Elastic Compute Cloud) Secondary Network resources.

## Example Usage

```terraform
list "aws_ec2_secondary_network" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
