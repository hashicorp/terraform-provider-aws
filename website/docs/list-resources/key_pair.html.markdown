---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_key_pair"
description: |-
  Lists EC2 Key Pair resources.
---

# List Resource: aws_key_pair

Lists EC2 Key Pair resources.

## Example Usage

```terraform
list "aws_key_pair" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
