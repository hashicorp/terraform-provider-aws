---
subcategory: "Transit Gateway"
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
  transit_gateway_multicast_domain_id = "tgw-mcast-domain-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `transit_gateway_multicast_domain_id` - (Optional) Identifier of the EC2 Transit Gateway Multicast Domain.

### filter Argument Reference

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayMulticastDomains.html).
* `values` - (Required) Set of values that are accepted for the given field. A multicast domain will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Multicast Domain identifier.
* `arn` - EC2 Transit Gateway Multicast Domain ARN.
* `associations` - EC2 Transit Gateway Multicast Domain Associations
    * `subnet_id` - The ID of the subnet associated with the transit gateway multicast domain.
    * `transit_gateway_attachment_id` - The ID of the transit gateway attachment.
* `auto_accept_shared_associations` - Whether to automatically accept cross-account subnet associations that are associated with the EC2 Transit Gateway Multicast Domain.
* `igmpv2_support` - Whether to enable Internet Group Management Protocol (IGMP) version 2 for the EC2 Transit Gateway Multicast Domain.
* `members` - EC2 Multicast Domain Group Members
    * `group_ip_address` - The IP address assigned to the transit gateway multicast group.
    * `network_interface_id` - The group members' network interface ID.
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway Multicast Domain.
* `sources` - EC2 Multicast Domain Group Sources
    * `group_ip_address` - The IP address assigned to the transit gateway multicast group.
    * `network_interface_id` - The group members' network interface ID.
* `static_sources_support` - Whether to enable support for statically configuring multicast group sources for the EC2 Transit Gateway Multicast Domain.
* `tags` - Key-value tags for the EC2 Transit Gateway Multicast Domain.
* `transit_gateway_id` - EC2 Transit Gateway identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
