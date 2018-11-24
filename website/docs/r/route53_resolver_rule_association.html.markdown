---
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule_association"
sidebar_current: "docs-aws-resource-route53-resolver-rule-association"
description: |-
  Provides a Route53 Resolver rule association.
---

# aws_route53_resolver_rule

Provides a Route53 Resolver rule association.

## Example Usage

```hcl
resource "aws_route53_resolver_rule_association" "example" {
  resolver_rule_id = "rslvr-rr-0123456789abcdef0"
  vpc_id = "vpc-01234567"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A name for the association that you're creating between a resolver rule and a VPC.
* `resolver_rule_id` - (Required) The ID of the resolver rule that you want to associate with the VPC.
* `vpc_id` - (Optional) The ID of the VPC that you want to associate the resolver rule with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resolver rule association.

## Import

Route53 Resolver rule associations can be imported using the `id`, e.g.

```
$ terraform import aws_route53_resolver_rule_association.example rslvr-rr-0123456789abcdef0
```
