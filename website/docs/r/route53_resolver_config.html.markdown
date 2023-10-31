---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_config"
description: |-
  Provides a Route 53 Resolver config resource.
---

# Resource: aws_route53_resolver_config

Provides a Route 53 Resolver config resource.

## Example Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_route53_resolver_config" "example" {
  resource_id              = aws_vpc.example.id
  autodefined_reverse_flag = "DISABLE"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_id` - (Required) The ID of the VPC that the configuration is for.
* `autodefined_reverse_flag` - (Required) Indicates whether or not the Resolver will create autodefined rules for reverse DNS lookups. Valid values: `ENABLE`, `DISABLE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the resolver configuration.
* `owner_id` - The AWS account ID of the owner of the VPC that this resolver configuration applies to.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Resolver configs using the Route 53 Resolver config ID. For example:

```terraform
import {
  to = aws_route53_resolver_config.example
  id = "rslvr-rc-715aa20c73a23da7"
}
```

Using `terraform import`, import Route 53 Resolver configs using the Route 53 Resolver config ID. For example:

```console
% terraform import aws_route53_resolver_config.example rslvr-rc-715aa20c73a23da7
```
