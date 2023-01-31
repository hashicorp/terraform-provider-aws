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

### Private NAT

```terraform
resource "aws_nat_gateway" "example" {
  connectivity_type = "private"
  subnet_id         = aws_subnet.example.id
}
```

## Argument Reference

The following arguments are supported:

* `allocation_id` - (Optional) The Allocation ID of the Elastic IP address for the gateway. Required for `connectivity_type` of `public`.
* `connectivity_type` - (Optional) Connectivity type for the gateway. Valid values are `private` and `public`. Defaults to `public`.
* `private_ip` - (Optional) The private IPv4 address to assign to the NAT gateway. If you don't provide an address, a private IPv4 address will be automatically assigned.
* `subnet_id` - (Required) The Subnet ID of the subnet in which to place the gateway.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the NAT Gateway.
* `network_interface_id` - The ID of the network interface associated with the NAT gateway.
* `public_ip` - The Elastic IP address associated with the NAT gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

NAT Gateways can be imported using the `id`, e.g.,

```
$ terraform import aws_nat_gateway.private_gw nat-05dba92075d71c408
```
