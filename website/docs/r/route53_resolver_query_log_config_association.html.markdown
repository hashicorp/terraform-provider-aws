---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_query_log_config_association"
description: |-
  Provides a Route 53 Resolver query logging configuration association resource.
---

# Resource: aws_route53_resolver_query_log_config_association

Provides a Route 53 Resolver query logging configuration association resource.

## Example Usage

```terraform
resource "aws_route53_resolver_query_log_config_association" "example" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.example.id
  resource_id                  = aws_vpc.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `resolver_query_log_config_id` - (Required) The ID of the [Route 53 Resolver query logging configuration](route53_resolver_query_log_config.html) that you want to associate a VPC with.
* `resource_id` - (Required) The ID of a VPC that you want this query logging configuration to log queries for.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` -The ID of the Route 53 Resolver query logging configuration association.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import  Route 53 Resolver query logging configuration associations using the Route 53 Resolver query logging configuration association ID. For example:

```terraform
import {
  to = aws_route53_resolver_query_log_config_association.example
  id = "rqlca-b320624fef3c4d70"
}
```

Using `terraform import`, import  Route 53 Resolver query logging configuration associations using the Route 53 Resolver query logging configuration association ID. For example:

```console
% terraform import aws_route53_resolver_query_log_config_association.example rqlca-b320624fef3c4d70
```
