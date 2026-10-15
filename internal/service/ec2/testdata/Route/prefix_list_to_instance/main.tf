# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_instance.test.primary_network_interface_id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  ipv6_address_count = 1
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = var.rName
}

data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = local.instance_types
  }

  preferred_instance_types = local.instance_types
}

locals {
  instance_types = ["t4g.nano", "t4g.micro"]
}

# acctest.ConfigVPCWithSubnetsIPv6(rName, 1)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true
}

# acctest.ConfigSubnetsIPv6(rName, 1)

resource "aws_subnet" "test" {
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

# acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("arm64")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-arm64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
