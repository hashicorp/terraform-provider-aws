---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_listener"
description: |-
  Lists ELB (Elastic Load Balancing) Listener resources.
---

# List Resource: aws_lb_listener

Lists ELB (Elastic Load Balancing) Listener resources.

## Example Usage

```terraform
list "aws_lb_listener" "example" {
  provider = aws

  config {
    load_balancer_arn = aws_lb.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `load_balancer_arn` - (Required) ARN of the Load Balancer to list Listeners from.
* `region` - (Optional) Region to query. Defaults to provider region.
