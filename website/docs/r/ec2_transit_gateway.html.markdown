---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway"
description: |-
  Manages an EC2 Transit Gateway
---

# Resource: aws_ec2_transit_gateway

Manages an EC2 Transit Gateway.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `amazon_side_asn` - (Optional) Private Autonomous System Number (ASN) for the Amazon side of a BGP session. The range is `64512` to `65534` for 16-bit ASNs and `4200000000` to `4294967294` for 32-bit ASNs. Default value: `64512`.

-> **NOTE:** Modifying `amazon_side_asn` on a Transit Gateway with active BGP sessions is [not allowed](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyTransitGatewayOptions.html). You must first delete all Transit Gateway attachments that have BGP configured prior to modifying `amazon_side_asn`.

* `auto_accept_shared_attachments` - (Optional) Whether resource attachment requests are automatically accepted. Valid values: `disable`, `enable`. Default value: `disable`.
* `default_route_table_association` - (Optional) Whether resource attachments are automatically associated with the default association route table. Valid values: `disable`, `enable`. Default value: `enable`.
* `default_route_table_propagation` - (Optional) Whether resource attachments automatically propagate routes to the default propagation route table. Valid values: `disable`, `enable`. Default value: `enable`.
* `description` - (Optional) Description of the EC2 Transit Gateway.
* `dns_support` - (Optional) Whether DNS support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.
* `security_group_referencing_support` - (Optional) Whether Security Group Referencing Support is enabled. Valid values: `disable`, `enable`. Default value: `disable`.
* `multicast_support` - (Optional) Whether Multicast support is enabled. Required to use `ec2_transit_gateway_multicast_domain`. Valid values: `disable`, `enable`. Default value: `disable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_cidr_blocks` - (Optional) One or more IPv4 or IPv6 CIDR blocks for the transit gateway. Must be a size /24 CIDR block or larger for IPv4, or a size /64 CIDR block or larger for IPv6.
* `vpn_ecmp_support` - (Optional) Whether VPN Equal Cost Multipath Protocol support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Amazon Resource Name (ARN)
* `association_default_route_table_id` - Identifier of the default association route table
* `id` - EC2 Transit Gateway identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway
* `propagation_default_route_table_id` - Identifier of the default propagation route table

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway` using the EC2 Transit Gateway identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway.example
  id = "tgw-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway` using the EC2 Transit Gateway identifier. For example:

```console
% terraform import aws_ec2_transit_gateway.example tgw-12345678
```
