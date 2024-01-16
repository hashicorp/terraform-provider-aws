---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_domain_list"
description: |-
  Provides a Route 53 Resolver DNS Firewall domain list resource.
---

# Resource: aws_route53_resolver_firewall_domain_list

Provides a Route 53 Resolver DNS Firewall domain list resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_domain_list" "example" {
  name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A name that lets you identify the domain list, to manage and use it.
* `domains` - (Optional) A array of domains for the firewall domain list.
* `tags` - (Optional) A map of tags to assign to the resource. f configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN (Amazon Resource Name) of the domain list.
* `id` - The ID of the domain list.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import  Route 53 Resolver DNS Firewall domain lists using the Route 53 Resolver DNS Firewall domain list ID. For example:

```terraform
import {
  to = aws_route53_resolver_firewall_domain_list.example
  id = "rslvr-fdl-0123456789abcdef"
}
```

Using `terraform import`, import  Route 53 Resolver DNS Firewall domain lists using the Route 53 Resolver DNS Firewall domain list ID. For example:

```console
% terraform import aws_route53_resolver_firewall_domain_list.example rslvr-fdl-0123456789abcdef
```
