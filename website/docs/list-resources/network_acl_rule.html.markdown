---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acl_rule"
description: |-
  Lists EC2 (Elastic Compute Cloud) Network ACL Rule resources.
---

# List Resource: aws_network_acl_rule

Lists EC2 Network ACL Rule resources.

## Example Usage

```terraform
list "aws_network_acl_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
