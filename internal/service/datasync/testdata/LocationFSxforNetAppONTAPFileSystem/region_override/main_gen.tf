# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  region = var.region

  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}

# testAccFSxOntapFileSystemConfig_baseNFS

resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  region = var.region

  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = var.rName
}

# testAccFSxOntapFileSystemConfig_base(rName, 1, "SINGLE_AZ_1")

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

resource "aws_fsx_ontap_file_system" "test" {
  region = var.region

  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test[0].id
}

# acctest.ConfigVPCWithSubnets(rName, 1)

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
  subnet_count = 1
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
