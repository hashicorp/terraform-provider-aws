# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

resource "aws_vpc" "example_vpc_1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-example-vpc-1"
  }
}

resource "aws_subnet" "example_subnet_1" {
  cidr_block = "10.1.0.0/24"
  vpc_id     = aws_vpc.example_vpc_1.id

  tags = {
    Name = "terraform-example-subnet-1"
  }
}

resource "aws_vpc" "example_vpc_2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-example-vpc-2"
  }
}

resource "aws_subnet" "example_subnet_2" {
  cidr_block = "10.2.0.0/24"
  vpc_id     = aws_vpc.example_vpc_2.id

  tags = {
    Name = "terraform-example-subnet-2"
  }
}

# Create the first Transit Gateway.
resource "aws_ec2_transit_gateway" "example_tgw_1" {
  tags = {
    Name = "terraform-example-tgw-1"
  }
}

# Attach the first VPC to the first Transit Gateway.
resource "aws_ec2_transit_gateway_vpc_attachment" "example_vpc_1_attachment" {
  subnet_ids         = [aws_subnet.example_subnet_1.id]
  transit_gateway_id = aws_ec2_transit_gateway.example_tgw_1.id
  vpc_id             = aws_vpc.example_vpc_1.id

  tags = {
    Name = "terraform-example-vpc-attach-1"
  }
}

# Create the second Transit Gateway in the same region.
resource "aws_ec2_transit_gateway" "example_tgw_2" {
  tags = {
    Name = "terraform-example-tgw-2"
  }
}

# Attach the second VPC to the second Transit Gateway.
resource "aws_ec2_transit_gateway_vpc_attachment" "example_vpc_2_attachment" {
  subnet_ids         = [aws_subnet.example_subnet_2.id]
  transit_gateway_id = aws_ec2_transit_gateway.example_tgw_2.id
  vpc_id             = aws_vpc.example_vpc_2.id

  tags = {
    Name = "terraform-example-vpc-attach-2"
  }
}

# Create the intra-region Peering Attachment from Gateway 1 to Gateway 2.
# Actually, this will create two peerings: one for Gateway 1 (Creator)
# and one for Gateway 2 (Acceptor).
resource "aws_ec2_transit_gateway_peering_attachment" "example_source_peering" {
  peer_region             = var.aws_region
  transit_gateway_id      = aws_ec2_transit_gateway.example_tgw_1.id
  peer_transit_gateway_id = aws_ec2_transit_gateway.example_tgw_2.id
  tags = {
    Name = "terraform-example-tgw-peering"
    Side = "Creator"
  }
}

# Transit Gateway 2's peering request needs to be accepted.
# So, we fetch the Peering Attachment that is created for the Gateway 2.
data "aws_ec2_transit_gateway_peering_attachment" "example_accepter_peering_data" {
  depends_on = [aws_ec2_transit_gateway_peering_attachment.example_source_peering]
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.example_tgw_2.id]
  }
}

# Accept the Attachment Peering request.
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "example_accepter" {
  transit_gateway_attachment_id = data.aws_ec2_transit_gateway_peering_attachment.example_accepter_peering_data.id
  tags = {
    Name = "terraform-example-tgw-peering-accepter"
    Side = "Acceptor"
  }
}