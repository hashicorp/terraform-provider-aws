---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule"
description: |-
  Provides a Route 53 Resolver DNS Firewall rule resource.
---

# Resource: aws_route53_resolver_firewall_rule

Provides a Route 53 Resolver DNS Firewall rule resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_domain_list" "example" {
  name    = "example"
  domains = ["example.com"]
  tags    = {}
}

resource "aws_route53_resolver_firewall_rule_group" "example" {
  name = "example"
  tags = {}
}

resource "aws_route53_resolver_firewall_rule" "example" {
  name                    = "example"
  action                  = "BLOCK"
  block_override_dns_type = "CNAME"
  block_override_domain   = "example.com"
  block_override_ttl      = 1
  block_response          = "OVERRIDE"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.example.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.example.id
  priority                = 100
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A name that lets you identify the rule, to manage and use it.
* `action` - (Required) The action that DNS Firewall should take on a DNS query when it matches one of the domains in the rule's domain list. Valid values: `ALLOW`, `BLOCK`, `ALERT`.
* `block_override_dns_type` - (Required if `block_response` is `OVERRIDE`) The DNS record's type. This determines the format of the record value that you provided in BlockOverrideDomain. Value values: `CNAME`.
* `block_override_domain` - (Required if `block_response` is `OVERRIDE`) The custom DNS record to send back in response to the query.
* `block_override_ttl` - (Required if `block_response` is `OVERRIDE`) The recommended amount of time, in seconds, for the DNS resolver or web browser to cache the provided override record. Minimum value of 0. Maximum value of 604800.
* `block_response` - (Required if `action` is `BLOCK`) The way that you want DNS Firewall to block the request. Valid values: `NODATA`, `NXDOMAIN`, `OVERRIDE`.
* `firewall_domain_list_id` - (Required) The ID of the domain list that you want to use in the rule.
* `firewall_domain_redirection_action` - (Optional) Evaluate DNS redirection in the DNS redirection chain, such as CNAME, DNAME, ot ALIAS. Valid values are `INSPECT_REDIRECTION_DOMAIN` and `TRUST_REDIRECTION_DOMAIN`. Default value is `INSPECT_REDIRECTION_DOMAIN`.
* `firewall_rule_group_id` - (Required) The unique identifier of the firewall rule group where you want to create the rule.
* `priority` - (Required) The setting that determines the processing order of the rule in the rule group. DNS Firewall processes the rules in a rule group by order of priority, starting from the lowest setting.
* `q_type` - (Optional) The query type you want the rule to evaluate. Additional details can be found [here](https://en.wikipedia.org/wiki/List_of_DNS_record_types)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the rule.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import  Route 53 Resolver DNS Firewall rules using the Route 53 Resolver DNS Firewall rule group ID and domain list ID separated by ':'. For example:

```terraform
import {
  to = aws_route53_resolver_firewall_rule.example
  id = "rslvr-frg-0123456789abcdef:rslvr-fdl-0123456789abcdef"
}
```

Using `terraform import`, import  Route 53 Resolver DNS Firewall rules using the Route 53 Resolver DNS Firewall rule group ID and domain list ID separated by ':'. For example:

```console
% terraform import aws_route53_resolver_firewall_rule.example rslvr-frg-0123456789abcdef:rslvr-fdl-0123456789abcdef
```
