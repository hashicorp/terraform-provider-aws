# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_docdbelastic_cluster" "test" {
  name           = var.rName
  shard_capacity = 2
  shard_count    = 1

  admin_user_name     = "testuser"
  admin_user_password = "testpassword"
  auth_type           = "PLAIN_TEXT"

  preferred_maintenance_window = "Tue:04:00-Tue:04:30"

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]

  subnet_ids = aws_subnet.test[*].id
}

# testAccClusterBaseConfig

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

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
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
