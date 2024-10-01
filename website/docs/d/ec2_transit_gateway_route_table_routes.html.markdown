---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_routes"
description: |-
   Provides informations for routes of a specific transit gateway
---

# Data Source: aws_ec2_transit_gateway_route_table_routes

Provides informations for routes of a specific transit gateway, such as state, type, cidr

## Example Usage

```terraform
data "aws_ec2_transit_gateway_route_table_routes" "test" {
  filter {
    name   = "type"
    values = ["propagated"]
  }
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

### Complexe use case with transit gateway peering

This example allow to create a mesh of transit gateway for différent regions routing all traffic to on-prem VPN

```terraform
resource "aws_ec2_transit_gateway" "this" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "this" {
  subnet_ids                                      = [for s in aws_subnet.private : s.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.this[0].id
  vpc_id                                          = aws_vpc.this.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
  depends_on                                      = [aws_subnet.private, aws_subnet.public]
}

resource "aws_ec2_transit_gateway_route_table" "this" {
  transit_gateway_id = local.my_transit_gateway_id
}

resource "aws_ec2_transit_gateway_route_table_association" "vpc" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.this.id
  transit_gateway_route_table_id = local.my_transit_gateway_id_route_table
}

resource "aws_ec2_transit_gateway_route_table_association" "vpn" {
  transit_gateway_attachment_id  = aws_vpn_connection.this[0].transit_gateway_attachment_id
  transit_gateway_route_table_id = local.my_transit_gateway_id_route_table
}

resource "aws_ec2_transit_gateway_route_table_propagation" "vpc" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.this.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.this.id
}

resource "aws_ec2_transit_gateway_route_table_propagation" "vpn" {
  transit_gateway_attachment_id  = aws_vpn_connection.this[0].transit_gateway_attachment_id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.this.id
}

provider "aws" {
  alias  = "eu-central-1"
  region = "eu-central-1"
}

resource "aws_ec2_transit_gateway" "eu-central-1" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
  provider                        = aws.eu-central-1
}

resource "aws_ec2_transit_gateway_peering_attachment" "eu-central-1" {
  peer_region             = "eu-central-1"
  peer_transit_gateway_id = aws_ec2_transit_gateway.eu-central-1.id
  transit_gateway_id      = aws_ec2_transit_gateway.this[0].id

  tags = {
    Name = "TGW mesh from eu-central-1"
  }
}

resource "aws_ec2_transit_gateway_route_table" "eu-central-1" {
  transit_gateway_id = aws_ec2_transit_gateway.eu-central-1.id
  tags               = merge({ Name = "wl-transit-gateway-routetable-eu-central-1" }, local.global_tags)
  provider           = aws.eu-central-1
}

resource "aws_ec2_transit_gateway_peering_attachment_accepter" "eu-central-1" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.eu-central-1.id
  provider                      = aws.eu-central-1
  tags = {
    Name = "Accepter TGW peering eu-central-1"
  }
}

data "aws_ec2_transit_gateway_vpc_attachments" "filtered-eu-central-1" {
  provider = aws.eu-central-1
  filter {
    name   = "state"
    values = ["pendingAcceptance", "available"]
  }
}

data "aws_ec2_transit_gateway_vpc_attachment" "unit-eu-central-1" {
  for_each = toset(data.aws_ec2_transit_gateway_vpc_attachments.filtered-eu-central-1.ids)
  id       = each.value
  provider = aws.eu-central-1
}

locals {
  trusted_aws_accounts_ids = {} # add to this list all account ids you trust

  trusted_vpc_attachments_list_eu-central-1 = compact([for k, tva in data.aws_ec2_transit_gateway_vpc_attachment.unit-eu-central-1 : contains(local.trusted_aws_accounts_ids, lookup(tva, "vpc_owner_id", "")) ? tva.id : ""])
  ## create a map with all vpc attachments trusted to be able to use for_each to avoid conflict on plan/apply ##
  trusted_vpc_attachements_eu-central-1 = toset(sort(local.trusted_vpc_attachments_list_eu-central-1))

}

resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "trusted_accounts_eu-central-1_tgw" {
  for_each                                        = local.trusted_vpc_attachements_eu-central-1
  transit_gateway_attachment_id                   = each.value
  transit_gateway_default_route_table_propagation = false
  transit_gateway_default_route_table_association = false
  tags                                            = local.global_tags
  provider                                        = aws.eu-central-1
  lifecycle {
    prevent_destroy = false
    ignore_changes  = [subnet_ids, id, dns_support, security_group_referencing_support, ipv6_support, transit_gateway_id, vpc_id, vpc_owner_id]
  }
}

resource "aws_ec2_transit_gateway_route_table_association" "trusted_accounts_eu-central-1" {
  for_each                       = aws_ec2_transit_gateway_vpc_attachment_accepter.trusted_accounts_eu-central-1_tgw
  transit_gateway_attachment_id  = each.value.transit_gateway_attachment_id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.eu-central-1.id
  provider                       = aws.eu-central-1
}

resource "aws_ec2_transit_gateway_route_table_propagation" "trusted_accounts_eu-central-1" {
  for_each                       = aws_ec2_transit_gateway_vpc_attachment_accepter.trusted_accounts_eu-central-1_tgw
  transit_gateway_attachment_id  = each.value.transit_gateway_attachment_id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.eu-central-1.id
  provider                       = aws.eu-central-1
}

data "aws_ec2_transit_gateway_route_table_routes" "test" {
  filter {
    name   = "type"
    values = ["propagated"]
  }
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.eu-central-1.id
  provider                       = aws.eu-central-1
  depends_on                     = [aws_ec2_transit_gateway_route_table_propagation.trusted_accounts_eu-central-1]
}

resource "aws_ec2_transit_gateway_route" "default-region-to-eu-central-1" {
  for_each                       = { for r in data.aws_ec2_transit_gateway_route_table_routes.test.routes : r.destination_cidr_block => r }
  destination_cidr_block         = each.key
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.this.id
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_peering_attachment.eu-central-1.id
}
```

## Argument Reference

The following arguments are required:

* `filter` - (Required) Custom filter block as described below.
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SearchTransitGatewayRoutes.html).
* `values` - (Required) Set of values that are accepted for the given field.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The transit gateway route table id suffixed by `-routes`
* `routes` - List of Transit Gateway Routes.

#### Routes list Attributes Reference

* `destination_cidr_block` - The CIDR used for route destination matches.
* `prefix_list_id` - The ID of the prefix list used for destination matches.
* `state` - The current state of the route, can be `active`, `deleted`, `pending`, `blackhole`, `deleting`.
* `transit_gateway_route_table_announcement_id` - The id of the transit gateway route table announcement, most of the time it is an empty string.
* `type` - The type of the route, can be `propagated` or `static`.
