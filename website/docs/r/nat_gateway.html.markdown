---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway"
description: |-
  Provides a resource to create a VPC NAT Gateway.
---

# Resource: aws_nat_gateway

Provides a resource to create a VPC NAT Gateway.

!> **WARNING:** You should not use the `aws_nat_gateway` resource that has `secondary_allocation_ids` in conjunction with an [`aws_nat_gateway_eip_association`](nat_gateway_eip_association.html) resource. Doing so may cause perpetual differences, and result in associations being overwritten.

## Example Usage

### Public NAT

```terraform
resource "aws_nat_gateway" "example" {
  allocation_id = aws_eip.example.id
  subnet_id     = aws_subnet.example.id

  tags = {
    Name = "gw NAT"
  }

  # To ensure proper ordering, it is recommended to add an explicit dependency
  # on the Internet Gateway for the VPC.
  depends_on = [aws_internet_gateway.example]
}
```

### Public NAT with Secondary Private IP Addresses

```terraform
resource "aws_nat_gateway" "example" {
  allocation_id                  = aws_eip.example.id
  subnet_id                      = aws_subnet.example.id
  secondary_allocation_ids       = [aws_eip.secondary.id]
  secondary_private_ip_addresses = ["10.0.1.5"]
}
```

### Private NAT

```terraform
resource "aws_nat_gateway" "example" {
  connectivity_type = "private"
  subnet_id         = aws_subnet.example.id
}
```

### Private NAT with Secondary Private IP Addresses

```terraform
resource "aws_nat_gateway" "example" {
  connectivity_type                  = "private"
  subnet_id                          = aws_subnet.example.id
  secondary_private_ip_address_count = 7
}
```

### Regional NAT Gateway with auto mode

```terraform
data "aws_availability_zones" "available" {}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "example" {
  vpc_id = aws_vpc.example.id
}

resource "aws_nat_gateway" "example" {
  vpc_id            = aws_vpc.example.id
  availability_mode = "regional"
}
```

### Regional NAT Gateway with manual mode

```terraform
data "aws_availability_zones" "available" {}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "example" {
  vpc_id = aws_vpc.example.id
}

resource "aws_eip" "example" {
  count  = 3
  domain = "vpc"
}

resource "aws_nat_gateway" "example" {
  vpc_id            = aws_vpc.example.id
  availability_mode = "regional"

  availability_zone_address {
    allocation_ids    = [aws_eip.example[0].id]
    availability_zone = data.aws_availability_zones.available.names[0]
  }
  availability_zone_address {
    allocation_ids    = [aws_eip.example[1].id, aws_eip.example[2].id]
    availability_zone = data.aws_availability_zones.available.names[1]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `allocation_id` - (Optional, zonal NAT gateways only) The Allocation ID of the Elastic IP address for the NAT Gateway. Required when `connectivity_type` is set to `public` and `availability_mode` is set to `zonal`. When `availability_mode` is set to `regional`, this must not be set; instead, use the `availability_zone_address` block to specify EIPs for each AZ.
* `availability_mode` - (Optional) Specifies whether to create a zonal (single-AZ) or regional (multi-AZ) NAT gateway. Valid values are `zonal` and `regional`. Defaults to `zonal`.
* `availability_zone_address` - (Optional, regional NAT gateways only) Repeatable configuration block for the Elastic IP addresses (EIPs) and availability zones for the regional NAT gateway. When not specified, the regional NAT gateway will automatically expand to new AZs and associate EIPs upon detection of an elastic network interface (auto mode). When specified, auto-expansion is disabled (manual mode). See [`availability_zone_address`](#availability_zone_address) below for details.
* `connectivity_type` - (Optional) Connectivity type for the NAT Gateway. Valid values are `private` and `public`. When `availability_mode` is set to `regional`, this must be set to `public`. Defaults to `public`.
* `private_ip` - (Optional, zonal NAT gateways only) The private IPv4 address to assign to the NAT Gateway. If you don't provide an address, a private IPv4 address will be automatically assigned.
* `subnet_id` - (Optional, zonal NAT gateways only) The Subnet ID of the subnet in which to place the NAT Gateway. Required when `availability_mode` is set to `zonal`. Must not be set when `availability_mode` is set to `regional`.
* `secondary_allocation_ids` - (Optional, zonal NAT gateways only) A list of secondary allocation EIP IDs for this NAT Gateway. To remove all secondary allocations an empty list should be specified.
* `secondary_private_ip_address_count` - (Optional, zonal and private NAT gateways only) The number of secondary private IPv4 addresses you want to assign to the NAT Gateway.
* `secondary_private_ip_addresses` - (Optional, zonal NAT gateways only) A list of secondary private IPv4 addresses to assign to the NAT Gateway. To remove all secondary private addresses an empty list should be specified.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Optional, regional NAT gateways only) VPC ID where this NAT Gateway will be created. Required when `availability_mode` is set to `regional`.

### `availability_zone_address`

~> **NOTE:** Once `availability_zone_address` blocks are specified (i.e., when using manual mode), removing `availability_zone_address` triggers recreation of the regional NAT Gateway in auto mode. Conversely, when operating in auto mode (i.e., without specifying `availability_zone_address`), adding these blocks triggers resource recreation to create a manual-mode regional NAT Gateway.

~> **NOTE:** Moving an `allocation_id` from one availability zone to another within `availability_zone_address` is not supported, because newly added EIPs are associated first, and only then are removed EIPs disassociated. To move it, remove the `allocation_id` from the source availability zone and apply the configuration. Then add it to the destination availability zone and apply again.

* `allocation_ids` - (Required) List of allocation IDs of the Elastic IP addresses (EIPs) to be used for handling outbound NAT traffic in this specific Availability Zone.
* `availability_zone` - (Optional) Availability Zone (e.g. `us-west-2a`) where this specific NAT gateway configuration will be active. Exactly one of `availability_zone` or `availability_zone_id` must be specified.
* `availability_zone_id` - (Optional) Availability Zone ID (e.g. `usw2-az2`) where this specific NAT gateway configuration will be active. Exactly one of `availability_zone` or `availability_zone_id` must be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `association_id` - (zonal NAT gateways only) The association ID of the Elastic IP address that's associated with the NAT Gateway. Only available when `connectivity_type` is `public`.
* `auto_provision_zones` - (regional NAT gateways only) Indicates whether AWS automatically manages AZ coverage.
* `auto_scaling_ips` - (regional NAT gateways only) Indicates whether AWS automatically allocates additional Elastic IP addresses (EIPs) in an AZ when the NAT gateway needs more ports due to increased concurrent connections to a single destination from that AZ.
* `id` - The ID of the NAT Gateway.
* `network_interface_id` - (zonal NAT gateways only) The ID of the network interface associated with the NAT Gateway.
* `public_ip` - (zonal NAT gateways only) The Elastic IP address associated with the NAT Gateway.
* `regional_nat_gateway_address` - (regional NAT gateways only) Repeatable blocks for information about the IP addresses and network interface associated with the regional NAT gateway.
    * `allocation_id` - Allocation ID of the Elastic IP address.
    * `association_id` - Association ID of the Elastic IP address.
    * `availability_zone` - Availability Zone where this specific NAT gateway configuration is active.
    * `availability_zone_id` - Availability Zone ID where this specific NAT gateway configuration is active
    * `network_interface_id` - ID of the network interface.
    * `public_ip` - Public IP address.
    * `status` - Status of the NAT gateway address.
* `route_table_id` - (regional NAT gateways only) ID of the automatically created route table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import NAT Gateways using the `id`. For example:

```terraform
import {
  to = aws_nat_gateway.private_gw
  id = "nat-05dba92075d71c408"
}
```

Using `terraform import`, import NAT Gateways using the `id`. For example:

```console
% terraform import aws_nat_gateway.private_gw nat-05dba92075d71c408
```
