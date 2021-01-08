---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_dnssec_config"
description: |-
  Provides a Route 53 Resolver DNSSEC config resource.
---

# Resource: aws_route53_resolver_dnssec_config

Provides a Route 53 Resolver DNSSEC config resource.

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_route53_resolver_dnssec_config" "example" {
  resource_id = aws_vpc.example.id
}
```

## Argument Reference

The following argument is supported:

* `resource_id` - (Required) The ID of the virtual private cloud (VPC) that you're updating the DNSSEC validation status for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID for a configuration for DNSSEC validation.
* `owner_id` - The owner account ID of the virtual private cloud (VPC) for a configuration for DNSSEC validation.
* `validation_status` - The validation status for a DNSSEC configuration. The status can be one of the following: `ENABLING`, `ENABLED`, `DISABLING` and `DISABLED`.

## Timeouts

`aws_route53_resolver_dnssec_config` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating Route 53 Resolver DNSSEC config
- `delete` - (Default `10 minutes`) Used for destroying Route 53 Resolver DNSSEC config

## Import

 Route 53 Resolver DNSSEC configs can be imported using the VPC ID, e.g.

```
$ terraform import aws_route53_resolver_dnssec_config.example vpc-7a190fdssf3
```
