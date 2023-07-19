---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule_association"
description: |-
  Provides a Route53 Resolver rule association.
---

# Resource: aws_route53_resolver_rule_association

Provides a Route53 Resolver rule association.

## Example Usage

```terraform
resource "aws_route53_resolver_rule_association" "example" {
  resolver_rule_id = aws_route53_resolver_rule.sys.id
  vpc_id           = aws_vpc.foo.id
}
```

## Argument Reference

This resource supports the following arguments:

* `resolver_rule_id` - (Required) The ID of the resolver rule that you want to associate with the VPC.
* `vpc_id` - (Required) The ID of the VPC that you want to associate the resolver rule with.
* `name` - (Optional) A name for the association that you're creating between a resolver rule and a VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the resolver rule association.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Resolver rule associations using the `id`. For example:

```terraform
import {
  to = aws_route53_resolver_rule_association.example
  id = "rslvr-rrassoc-97242eaf88example"
}
```

Using `terraform import`, import Route53 Resolver rule associations using the `id`. For example:

```console
% terraform import aws_route53_resolver_rule_association.example rslvr-rrassoc-97242eaf88example
```
