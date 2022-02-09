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
  vpc_id     = aws_vpc.vpc1.id
  cidr_block = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet2" {
  vpc_id     = aws_vpc.vpc1.id
  cidr_block = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet3" {
  vpc_id     = aws_vpc.vpc2.id
  cidr_block = "11.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_subnet" "subnet4" {
  vpc_id     = aws_vpc.vpc2.id
  cidr_block = "11.0.2.0/24"
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
    subnet_ids                    = [aws_subnet.subnet3.id,aws_subnet.subnet4.id]
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

* `transit_gateway_id` - (Required, Forces new resource) EC2 Transit Gateway Identifier. The target resource must have `multicast_support = "enable"`.
* `association` - (Optional) Can be specified multiple times for different EC2 Transit Gateway Attachments. Each association block supports the fields documented below. This argument is processed in [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `members` - (Optional) Can be specified multiple times for different Group IP Addresses. Each members block supports the fields documented below. This argument is processed in [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `sources` - (Optional) Can be specified multiple times for different Group IP Addresses. Each members block supports the fields documented below. This argument is processed in [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `association` block supports:

* `transit_gateway_attachment_id` - (Required) EC2 Transit Gateway Attachment Identifier.
* `subnet_ids` - (Required, Minimum items: 1) List of subnets identifiers to associate. The listed subnets must reside within the specified EC2 Transit Gateway Attachment.

The `members` and `sources` blocks support:

* `group_ip_address` - (Required) Multicast Group IP address. Must be valid IPv4 or IPv6 IP Address in the 224.0.0.0/4 or ff00::/8 CIDR range.
* `network_interface_ids` - (Required, Minimum items: 1) List of Network Interface Identifiers to create Multicast Group for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Multicast Domain identifier

## Import

`aws_ec2_transit_gateway_multicast_domain` can be imported by using the EC2 Transit Gateway Multicast Domain identifier, e.g.,

```
terraform import aws_ec2_transit_gateway_multicast_domain.example tgw-mcast-domain-12345
```