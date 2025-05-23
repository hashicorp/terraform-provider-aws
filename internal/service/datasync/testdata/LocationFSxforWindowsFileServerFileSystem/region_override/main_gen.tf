# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_datasync_location_fsx_windows_file_system" "test" {
  region = var.region

  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]
}

# testAccLocationFSxWindowsFileSystemConfig_baseFS

resource "aws_fsx_windows_file_system" "test" {
  region = var.region

  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}

resource "aws_security_group" "test" {
  region = var.region

  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

# testAccLocationFSxWindowsFileSystemConfig_base

resource "aws_directory_service_directory" "test" {
  region = var.region

  edition  = "Standard"
  name     = var.domain
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  region = var.region

  count = local.subnet_count

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

locals {
  subnet_count = 2
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "domain" {
  type     = string
  nullable = false
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
