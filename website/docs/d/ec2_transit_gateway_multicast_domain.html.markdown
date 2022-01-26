---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_domain"
description: |-
  Get information on an EC2 Transit Gateway Multicast Domain
---

# Data Source: aws_ec2_transit_gateway_multicast_domain

Get information on an EC2 Transit Gateway Multicast Domain.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway_multicast_domain" "example" {
  filter {
    name   = "transit-gateway-multicast-domain-id"
    values = ["tgw-mcast-domain-12345678"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway_multicast_domain" "example" {
  id = "tgw-mcast-domain-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway Route Table.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Route Table Amazon Resource Name (ARN).
* `id` - EC2 Transit Gateway Route Table identifier
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway Route Table
* `igmpv2_support` - Whether IGMPv2 is supported
* `static_source_support` -  Whether Static Sources are supported
* `auto_accept_shared_associations` -  Whether Shared Associations are Auto Accepted
