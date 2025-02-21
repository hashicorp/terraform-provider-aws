---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_endpoint"
description: |-
  Provides a Route 53 Resolver endpoint resource.
---

# Resource: aws_route53_resolver_endpoint

Provides a Route 53 Resolver endpoint resource.

## Example Usage

```terraform
resource "aws_route53_resolver_endpoint" "foo" {
  name                   = "foo"
  direction              = "INBOUND"
  resolver_endpoint_type = "IPV4"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
    ip        = "10.0.64.4"
  }

  protocols = ["Do53", "DoH"]

  tags = {
    Environment = "Prod"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `direction` - (Required) Direction of DNS queries to or from the Route 53 Resolver endpoint.
Valid values are `INBOUND` (resolver forwards DNS queries to the DNS service for a VPC from your network or another VPC)
or `OUTBOUND` (resolver forwards DNS queries from the DNS service for a VPC to your network or another VPC).
* `ip_address` - (Required) Subnets and IP addresses in your VPC that you want DNS queries to pass through on the way from your VPCs
to your network (for outbound endpoints) or on the way from your network to your VPCs (for inbound endpoints). Described below.
* `name` - (Optional) Friendly name of the Route 53 Resolver endpoint.
* `protocols` - (Optional) Protocols you want to use for the Route 53 Resolver endpoint.
Valid values are `DoH`, `Do53`, or `DoH-FIPS`.
* `resolver_endpoint_type` - (Optional) Endpoint IP type. This endpoint type is applied to all IP addresses.
Valid values are `IPV6`,`IPV4` or `DUALSTACK` (both IPv4 and IPv6).
* `security_group_ids` - (Required) ID of one or more security groups that you want to use to control access to this VPC.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `ip_address` object supports the following:

* `ip` - (Optional) IPv4 address in the subnet that you want to use for DNS queries.
* `ipv6` - (Optional) IPv6 address in the subnet that you want to use for DNS queries.
* `subnet_id` - (Required) ID of the subnet that contains the IP address.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Route 53 Resolver endpoint.
* `host_vpc_id` - ID of the VPC that you want to create the resolver endpoint in.
* `id` - ID of the Route 53 Resolver endpoint.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import  Route 53 Resolver endpoints using the Route 53 Resolver endpoint ID. For example:

```terraform
import {
  to = aws_route53_resolver_endpoint.foo
  id = "rslvr-in-abcdef01234567890"
}
```

Using `terraform import`, import  Route 53 Resolver endpoints using the Route 53 Resolver endpoint ID. For example:

```console
% terraform import aws_route53_resolver_endpoint.foo rslvr-in-abcdef01234567890
```
