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
  name      = "foo"
  direction = "INBOUND"

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

* `direction` - (Required) The direction of DNS queries to or from the Route 53 Resolver endpoint.
Valid values are `INBOUND` (resolver forwards DNS queries to the DNS service for a VPC from your network or another VPC)
or `OUTBOUND` (resolver forwards DNS queries from the DNS service for a VPC to your network or another VPC).
* `ip_address` - (Required) The subnets and IP addresses in your VPC that you want DNS queries to pass through on the way from your VPCs
to your network (for outbound endpoints) or on the way from your network to your VPCs (for inbound endpoints). Described below.
* `security_group_ids` - (Required) The ID of one or more security groups that you want to use to control access to this VPC.
* `name` - (Optional) The friendly name of the Route 53 Resolver endpoint.
* `protocols` - (Optional) The protocols you want to use for the Route 53 Resolver endpoint. Valid values: `DoH`, `Do53`, `DoH-FIPS`.
* `resolver_endpoint_type` - (Optional) The Route 53 Resolver endpoint IP address type. Valid values: `IPV4`, `IPV6`, `DUALSTACK`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `ip_address` object supports the following:

* `subnet_id` - (Required) The ID of the subnet that contains the IP address.
* `ip` - (Optional) The IP address in the subnet that you want to use for DNS queries.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Route 53 Resolver endpoint.
* `arn` - The ARN of the Route 53 Resolver endpoint.
* `host_vpc_id` - The ID of the VPC that you want to create the resolver endpoint in.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
