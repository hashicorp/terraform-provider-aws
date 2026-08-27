---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_key_pair"
description: |-
  Lists EC2 (Elastic Compute Cloud) Key Pair resources.
---

# List Resource: aws_ec2_key_pair

Lists EC2 (Elastic Compute Cloud) Key Pair resources.

## Example Usage

```terraform
list "aws_ec2_key_pair" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
