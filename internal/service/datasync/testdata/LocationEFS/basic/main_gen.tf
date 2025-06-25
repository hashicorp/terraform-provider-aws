# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test[0].arn
  }
}

# testAccLocationEFSConfig_base

#resource "aws_vpc" "test" {
#  cidr_block = "10.0.0.0/16"
#}
#
#resource "aws_subnet" "test" {
#  cidr_block = "10.0.0.0/24"
#  vpc_id     = aws_vpc.test.id
#}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_efs_file_system" "test" {
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
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

