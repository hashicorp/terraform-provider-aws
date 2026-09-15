---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_listener_rule"
description: |-
  Lists ELB (Elastic Load Balancing) Listener Rule resources.
---

# List Resource: aws_lb_listener_rule

Lists ELB (Elastic Load Balancing) Listener Rule resources.

The default Listener Rule is managed by the `aws_lb_listener` resource,
so it is not returned by the `aws_lb_listener_rule` List Resource.

## Example Usage

```terraform
list "aws_lb_listener_rule" "example" {
  provider = aws

  config {
    listener_arn = aws_lb_listener.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `listener_arn` - (Required) ARN of the Listener to list ListenerRules from.
* `region` - (Optional) Region to query. Defaults to provider region.
