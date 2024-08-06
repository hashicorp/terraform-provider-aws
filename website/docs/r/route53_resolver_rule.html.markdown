---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule"
description: |-
  Provides a Route53 Resolver rule.
---

# Resource: aws_route53_resolver_rule

Provides a Route53 Resolver rule.

## Example Usage

### System rule

```terraform
resource "aws_route53_resolver_rule" "sys" {
  domain_name = "subdomain.example.com"
  rule_type   = "SYSTEM"
}
```

### Forward rule

```terraform
resource "aws_route53_resolver_rule" "fwd" {
  domain_name          = "example.com"
  name                 = "example"
  rule_type            = "FORWARD"
  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "123.45.67.89"
  }

  tags = {
    Environment = "Prod"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) DNS queries for this domain name are forwarded to the IP addresses that are specified using `target_ip`.
* `rule_type` - (Required) The rule type. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`.
* `name` - (Optional) A friendly name that lets you easily find a rule in the Resolver dashboard in the Route 53 console.
* `resolver_endpoint_id` (Optional) The ID of the outbound resolver endpoint that you want to use to route DNS queries to the IP addresses that you specify using `target_ip`.
This argument should only be specified for `FORWARD` type rules.
* `target_ip` - (Optional) Configuration block(s) indicating the IPs that you want Resolver to forward DNS queries to (documented below).
This argument should only be specified for `FORWARD` type rules.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `target_ip` object supports the following:

* `ip` - (Required) One IP address that you want to forward DNS queries to. You can specify only IPv4 addresses.
* `port` - (Optional) The port at `ip` that you want to forward DNS queries to. Default value is `53`.
* `protocol` - (Optional) The protocol for the resolver endpoint. Valid values can be found in the [AWS documentation](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53resolver_TargetAddress.html). Default value is `Do53`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the resolver rule.
* `arn` - The ARN (Amazon Resource Name) for the resolver rule.
* `owner_id` - When a rule is shared with another AWS account, the account ID of the account that the rule is shared with.
* `share_status` - Whether the rules is shared and, if so, whether the current account is sharing the rule with another account, or another account is sharing the rule with the current account.
Values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Resolver rules using the `id`. For example:

```terraform
import {
  to = aws_route53_resolver_rule.sys
  id = "rslvr-rr-0123456789abcdef0"
}
```

Using `terraform import`, import Route53 Resolver rules using the `id`. For example:

```console
% terraform import aws_route53_resolver_rule.sys rslvr-rr-0123456789abcdef0
```
