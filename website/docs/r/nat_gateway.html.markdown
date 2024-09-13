---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway"
description: |-
  Provides a resource to create a VPC NAT Gateway.
---

# Resource: aws_nat_gateway

Provides a resource to create a VPC NAT Gateway.

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

## Argument Reference

This resource supports the following arguments:

* `allocation_id` - (Optional) The Allocation ID of the Elastic IP address for the NAT Gateway. Required for `connectivity_type` of `public`.
* `connectivity_type` - (Optional) Connectivity type for the NAT Gateway. Valid values are `private` and `public`. Defaults to `public`.
* `private_ip` - (Optional) The private IPv4 address to assign to the NAT Gateway. If you don't provide an address, a private IPv4 address will be automatically assigned.
* `subnet_id` - (Required) The Subnet ID of the subnet in which to place the NAT Gateway.
* `secondary_allocation_ids` - (Optional) A list of secondary allocation EIP IDs for this NAT Gateway.
* `secondary_private_ip_address_count` - (Optional) [Private NAT Gateway only] The number of secondary private IPv4 addresses you want to assign to the NAT Gateway.
* `secondary_private_ip_addresses` - (Optional) A list of secondary private IPv4 addresses to assign to the NAT Gateway.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `association_id` - The association ID of the Elastic IP address that's associated with the NAT Gateway. Only available when `connectivity_type` is `public`.
* `id` - The ID of the NAT Gateway.
* `network_interface_id` - The ID of the network interface associated with the NAT Gateway.
* `public_ip` - The Elastic IP address associated with the NAT Gateway.
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
