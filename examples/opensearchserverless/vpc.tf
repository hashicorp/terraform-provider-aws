# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Creates a VPC
resource "aws_vpc" "vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "example-vpc"
  }
}

# Creates a subnet
resource "aws_subnet" "subnet" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example-subnet"
  }
}

# Creates an Internet Gateway
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "example-igw"
  }
}

# Creates a route table
resource "aws_route_table" "route_table" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }
}

# Creates a route table association
resource "aws_route_table_association" "route_table_association" {
  subnet_id      = aws_subnet.subnet.id
  route_table_id = aws_route_table.route_table.id
}

# Creates a security group
resource "aws_security_group" "security_group" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "example-security-group"
  }
}

# Allows all outbound traffic
resource "aws_vpc_security_group_egress_rule" "sg_egress" {
  security_group_id = aws_security_group.security_group.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}

# Allows inbound traffic from within security group
resource "aws_vpc_security_group_ingress_rule" "sg_ingress" {
  security_group_id = aws_security_group.security_group.id

  referenced_security_group_id = aws_security_group.security_group.id
  ip_protocol                  = "-1"
}
