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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
