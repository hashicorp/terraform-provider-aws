---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_group"
description: |-
  Lists ELB (Elastic Load Balancing) Target Group resources.
---

# List Resource: aws_lb_target_group

Lists ELB (Elastic Load Balancing) Target Group resources.

## Example Usage

```terraform
list "aws_lb_target_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
