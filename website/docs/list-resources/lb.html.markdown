---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb"
description: |-
  Lists ELB (Elastic Load Balancing) Load Balancer resources.
---

# List Resource: aws_lb

Lists ELB (Elastic Load Balancing) Load Balancer resources.

## Example Usage

```terraform
list "aws_lb" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
