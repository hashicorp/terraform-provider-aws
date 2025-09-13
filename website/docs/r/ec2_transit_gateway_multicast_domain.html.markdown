---
subcategory: "Transit Gateway"
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
}

resource "aws_vpc" "vpc2" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "subnet1" {
  vpc_id            = aws_vpc.vpc1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "subnet2" {
  vpc_id            = aws_vpc.vpc1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

resource "aws_subnet" "subnet3" {
  vpc_id            = aws_vpc.vpc2.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_instance" "instance1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet1.id
}

resource "aws_instance" "instance2" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet2.id
}

resource "aws_instance" "instance3" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet3.id
}

resource "aws_ec2_transit_gateway" "tgw" {
  multicast_support = "enable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "attachment1" {
  subnet_ids         = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc1.id
}

resource "aws_ec2_transit_gateway_vpc_attachment" "attachment2" {
  subnet_ids         = [aws_subnet.subnet3.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc2.id
}

resource "aws_ec2_transit_gateway_multicast_domain" "domain" {
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id

  static_sources_support = "enable"

  tags = {
    Name = "Transit_Gateway_Multicast_Domain_Example"
  }
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "association3" {
  subnet_id                           = aws_subnet.subnet3.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.attachment2.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.domain.id
}

resource "aws_ec2_transit_gateway_multicast_group_source" "source" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_instance.instance3.primary_network_interface_id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.association3.transit_gateway_multicast_domain_id
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "association1" {
  subnet_id                           = aws_subnet.subnet1.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.attachment1.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.domain.id
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "association2" {
  subnet_id                           = aws_subnet.subnet2.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.attachment2.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.domain.id
}

resource "aws_ec2_transit_gateway_multicast_group_member" "member1" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_instance.instance1.primary_network_interface_id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.association1.transit_gateway_multicast_domain_id
}

resource "aws_ec2_transit_gateway_multicast_group_member" "member2" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_instance.instance2.primary_network_interface_id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.association1.transit_gateway_multicast_domain_id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_id` - (Required) EC2 Transit Gateway identifier. The EC2 Transit Gateway must have `multicast_support` enabled.
* `auto_accept_shared_associations` - (Optional) Whether to automatically accept cross-account subnet associations that are associated with the EC2 Transit Gateway Multicast Domain. Valid values: `disable`, `enable`. Default value: `disable`.
* `igmpv2_support` - (Optional) Whether to enable Internet Group Management Protocol (IGMP) version 2 for the EC2 Transit Gateway Multicast Domain. Valid values: `disable`, `enable`. Default value: `disable`.
* `static_sources_support` - (Optional) Whether to enable support for statically configuring multicast group sources for the EC2 Transit Gateway Multicast Domain. Valid values: `disable`, `enable`. Default value: `disable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Multicast Domain. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Multicast Domain identifier.
* `arn` - EC2 Transit Gateway Multicast Domain Amazon Resource Name (ARN).
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway Multicast Domain.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_multicast_domain` using the EC2 Transit Gateway Multicast Domain identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_multicast_domain.example
  id = "tgw-mcast-domain-12345"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_multicast_domain` using the EC2 Transit Gateway Multicast Domain identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_multicast_domain.example tgw-mcast-domain-12345
```
