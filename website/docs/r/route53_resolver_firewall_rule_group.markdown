---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group"
description: |-
  Provides a Route 53 Resolver DNS Firewall rule group resource.
---

# Resource: aws_route53_resolver_firewall_rule_group

Provides a Route 53 Resolver DNS Firewall rule group resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_rule_group" "example" {
  name = "example"
}
```

## Argument Reference

The following argument is supported:

* `name` - (Required) A name that lets you identify the rule group, to manage and use it.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN (Amazon Resource Name) of the rule group.
* `id` - The ID of the rule group.

## Import

 Route 53 Resolver DNS Firewall rule groups can be imported using the Route 53 Resolver DNS Firewall rule group ID, e.g.

```
$ terraform import aws_route53_resolver_firewall_rule_group.example rslvr-frg-0123456789abcdef
```
