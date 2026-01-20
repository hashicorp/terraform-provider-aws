---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_security_group"
description: |-
  Lists EC2 (Elastic Compute Cloud) Security Group resources.
---

# List Resource: aws_security_group

Lists EC2 (Elastic Compute Cloud) Security Group resources.

## Example Usage

```terraform
list "aws_security_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
