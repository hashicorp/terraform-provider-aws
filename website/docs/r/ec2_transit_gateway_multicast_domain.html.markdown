---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_domain"
description: |-
  Manages an EC2 Transit Gateway Multicast Domain
---

# Resource: aws_ec2_transit_gateway_multicast_domain

Manages an EC2 Transit Gateway Multicast Domain.

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}

resource "aws_vpc" "vpc1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_vpc" "vpc2" {
  cidr_block = "11.0.0.0/16"
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet1" {
  vpc_id            = aws_vpc.vpc1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet2" {
  vpc_id            = aws_vpc.vpc1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet3" {
  vpc_id            = aws_vpc.vpc2.id
  cidr_block        = "11.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet4" {
  vpc_id            = aws_vpc.vpc2.id
  cidr_block        = "11.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_instance" "instance1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet1.id
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_instance" "instance2" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet2.id
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_ec2_transit_gateway" "tgw" {
  multicast_support = "enable"
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "attachment1" {
  subnet_ids         = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc1.id
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test2" {
  subnet_ids         = [aws_subnet.subnet3.id, aws_subnet.subnet4.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc2.id
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "multicast_domain" {
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.attachment1.id
    subnet_ids                    = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]
  }
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.attachment2.id
    subnet_ids                    = [aws_subnet.subnet3.id, aws_subnet.subnet4.id]
  }
  members {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.instance1.primary_network_interface_id]
  }
  members {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.instancte2.primary_network_interface_id]
  }
  sources {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.instance1.primary_network_interface_id]
  }
  sources {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.instance2.primary_network_interface_id]
  }
  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required) EC2 Transit Gateway identifier. The EC2 Transit Gateway must have `multicast_support` enabled.
* `auto_accept_shared_associations` - (Optional) Whether to automatically accept cross-account subnet associations that are associated with the EC2 Transit Gateway Multicast Domain. Valid values: `disable`, `enable`. Default value: `disable`.
* `igmpv2_support` - (Optional) Whether to enable Internet Group Management Protocol (IGMP) version 2 for the EC2 Transit Gateway Multicast Domain. Valid values: `disable`, `enable`. Default value: `disable`.
* `static_sources_support` - (Optional) Whether to enable support for statically configuring multicast group sources for the EC2 Transit Gateway Multicast. Valid values: `disable`, `enable`. Default value: `disable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Multicast Domain. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Multicast Domain identifier.
* `arn` - EC2 Transit Gateway Multicast Domain Amazon Resource Name (ARN).
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway Multicast Domain.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_ec2_transit_gateway_multicast_domain` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for multicast domain creation
- `delete` - (Default `10 minutes`) Used for multicast domain deletion

## Import

`aws_ec2_transit_gateway_multicast_domain` can be imported by using the EC2 Transit Gateway Multicast Domain identifier, e.g.,

```
terraform import aws_ec2_transit_gateway_multicast_domain.example tgw-mcast-domain-12345
```