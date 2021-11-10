---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway"
description: |-
  Get information on an EC2 Transit Gateway
---

# Data Source: aws_ec2_transit_gateway

Get information on an EC2 Transit Gateway.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway" "example" {
  filter {
    name   = "options.amazon-side-asn"
    values = ["64512"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway" "example" {
  id = "tgw-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

#### Supported values for Filters

One or more filters are supported. The possible values are:

* `options.propagation-default-route-table-id` - The ID of the default propagation route table.
* `options.amazon-side-asn` - The private ASN for the Amazon side of a BGP session.
* `options.association-default-route-table-id` - The ID of the default association route table.
* `options.auto-accept-shared-attachments` - Indicates whether there is automatic acceptance of attachment requests (enable | disable ).
* `options.default-route-table-association` - Indicates whether resource attachments are automatically associated with the default association route table (enable | disable ).
* `options.default-route-table-propagation` - Indicates whether resource attachments automatically propagate routes to the default propagation route table (enable | disable ).
* `options.dns-support` - Indicates whether DNS support is enabled (enable | disable ).
* `options.vpn-ecmp-support` - Indicates whether Equal Cost Multipath Protocol support is enabled (enable | disable ).
* `owner-id` - The ID of the Amazon Web Services account that owns the transit gateway.
* `state` - The state of the transit gateway (available | deleted | deleting | modifying | pending ).
* `transit-gateway-id` - The ID of the transit gateway

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `amazon_side_asn` - Private Autonomous System Number (ASN) for the Amazon side of a BGP session
* `arn` - EC2 Transit Gateway Amazon Resource Name (ARN)
* `association_default_route_table_id` - Identifier of the default association route table
* `auto_accept_shared_attachments` - Whether resource attachment requests are automatically accepted.
* `default_route_table_association` - Whether resource attachments are automatically associated with the default association route table.
* `default_route_table_propagation` - Whether resource attachments automatically propagate routes to the default propagation route table.
* `description` - Description of the EC2 Transit Gateway
* `dns_support` - Whether DNS support is enabled.
* `id` - EC2 Transit Gateway identifier
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway
* `propagation_default_route_table_id` - Identifier of the default propagation route table.
* `tags` - Key-value tags for the EC2 Transit Gateway
* `vpn_ecmp_support` - Whether VPN Equal Cost Multipath Protocol support is enabled.
