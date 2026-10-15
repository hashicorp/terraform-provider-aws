---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_elb"
description: |-
  Lists ELB Classic Load Balancer resources.
---

# List Resource: aws_elb

Lists ELB Classic Load Balancer resources.

## Example Usage

```terraform
list "aws_elb" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
