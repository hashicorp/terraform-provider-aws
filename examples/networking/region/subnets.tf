# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_availability_zones" "all" {}

module "primary_subnet" {
  source            = "../subnet"
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.all.names[0]
}

module "secondary_subnet" {
  source            = "../subnet"
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.all.names[1]
}
