---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_flow_log"
description: |-
  Lists VPC Flow Log resources.
---

# List Resource: aws_flow_log

Lists VPC Flow Log resources.

## Example Usage

```terraform
list "aws_flow_log" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
