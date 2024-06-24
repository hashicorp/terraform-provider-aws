---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_config"
description: |-
  Provides a Route 53 Resolver DNS Firewall config resource.
---

# Resource: aws_route53_resolver_firewall_config

Provides a Route 53 Resolver DNS Firewall config resource.

## Example Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_route53_resolver_firewall_config" "example" {
  resource_id        = aws_vpc.example.id
  firewall_fail_open = "ENABLED"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_id` - (Required) The ID of the VPC that the configuration is for.
* `firewall_fail_open` - (Required) Determines how Route 53 Resolver handles queries during failures, for example when all traffic that is sent to DNS Firewall fails to receive a reply. By default, fail open is disabled, which means the failure mode is closed. This approach favors security over availability. DNS Firewall blocks queries that it is unable to evaluate properly. If you enable this option, the failure mode is open. This approach favors availability over security. DNS Firewall allows queries to proceed if it is unable to properly evaluate them. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the firewall configuration.
* `owner_id` - The AWS account ID of the owner of the VPC that this firewall configuration applies to.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Resolver DNS Firewall configs using the Route 53 Resolver DNS Firewall config ID. For example:

```terraform
import {
  to = aws_route53_resolver_firewall_config.example
  id = "rdsc-be1866ecc1683e95"
}
```

Using `terraform import`, import Route 53 Resolver DNS Firewall configs using the Route 53 Resolver DNS Firewall config ID. For example:

```console
% terraform import aws_route53_resolver_firewall_config.example rdsc-be1866ecc1683e95
```
