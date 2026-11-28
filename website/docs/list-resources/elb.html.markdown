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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
