# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_datasync_location_nfs" "test" {
  region = var.region

  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}

# testAccLocationNFSConfig_base

resource "aws_datasync_agent" "test" {
  region = var.region

  ip_address = aws_instance.test.public_ip
  name       = var.rName
}

# testAccAgentAgentConfig_base

# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  region = var.region

  name = "/aws/service/datasync/ami"
}

resource "aws_internet_gateway" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  region = var.region

  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  region = var.region

  name   = var.rName
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  region = var.region

  depends_on = [aws_internet_gateway.test]

  ami                    = data.aws_ssm_parameter.aws_service_datasync_ami.value
  instance_type          = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test[0].id

  # The Instance must have a public IP address because the aws_datasync_agent retrieves
  # the activation key by making an HTTP request to the instance
  associate_public_ip_address = true
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  region = var.region

  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  region = var.region

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

# acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "m5.2xlarge", "m5.4xlarge")

data "aws_ec2_instance_type_offering" "available" {
  region = var.region

  filter {
    name   = "instance-type"
    values = local.preferred_instance_types
  }

  filter {
    name   = "location"
    values = [aws_subnet.test[0].availability_zone]
  }

  location_type            = "availability-zone"
  preferred_instance_types = local.preferred_instance_types
}

locals {
  preferred_instance_types = ["m5.2xlarge", "m5.4xlarge"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
