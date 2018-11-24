---
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule"
sidebar_current: "docs-aws-resource-route53-resolver-rule"
description: |-
  Provides a Route53 Resolver rule.
---

# aws_route53_resolver_rule

Provides a Route53 Resolver rule.

## Example Usage

```hcl
resource "aws_route53_resolver_rule" "example" {
  domain_name          = "example.com"
  name                 = "example"
  resolver_endpoint_id = "..."
  rule_type            = "FORWARD"

  target_ip {
    ip   = "123.45.67.89"
    port = 1234
  }

  tags {
    Foo = "Barr"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) DNS queries for this domain name are forwarded to the IP addresses that are specified using `target_ip`.
* `name` - (Optional) A friendly name that lets you easily find a rule in the Resolver dashboard in the Route 53 console.
* `resolver_endpoint_id` (Optional) The ID of the outbound resolver endpoint that you want to use to route DNS queries to the IP addresses that you specify using `target_ip`.
* `rule_type` - (Required) The rule type. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`.
* `target_ip` - (Optional) Configuration block(s) indicating the IPs that you want Resolver to forward DNS queries to (documented below).
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `target_ip` object supports the following:

* `ip` - (Required) One IP address that you want to forward DNS queries to. You can specify only IPv4 addresses.
* `port` - (Optional) The port at `ip` that you want to forward DNS queries to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resolver rule.
* `arn` - The ARN (Amazon Resource Name) for the resolver rule.

## Import

Route53 Resolver rules can be imported using the `id`, e.g.

```
$ terraform import aws_route53_resolver_rule.example rslvr-rr-0123456789abcdef0
```
